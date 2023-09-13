use std::env;
use std::error::Error;
use std::net;

use axum::{
    extract::State,
    http::{StatusCode, Request, HeaderValue},
    routing::{get, post},
    Json, Router, body::Body, response::IntoResponse,
};
use tower_http::validate_request::ValidateRequestHeaderLayer;

use clap::{Parser, Subcommand};
use ndc_client::models::{
    CapabilitiesResponse, ErrorResponse, ExplainResponse, MutationRequest, MutationResponse,
    QueryRequest, QueryResponse, SchemaResponse,
};
use opentelemetry::{global, sdk::propagation::TraceContextPropagator};
use opentelemetry_api::KeyValue;
use opentelemetry_otlp::{WithExportConfig, OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT};
use opentelemetry_sdk::trace::Sampler;
use prometheus::Registry;
use schemars::{schema::RootSchema, JsonSchema};
use serde::{de::DeserializeOwned, Serialize};
use tower_http::{
    cors::CorsLayer,
    trace::{DefaultMakeSpan, TraceLayer},
};
use tracing::Level;
use tracing_subscriber::{prelude::*, EnvFilter};

use crate::check_health;
use crate::connector::{Connector, InvalidRange, SchemaError, UpdateConfigurationError};
use crate::routes;

#[derive(Parser)]
struct CliArgs {
    #[command(subcommand)]
    command: Command,
}

#[derive(Clone, Subcommand)]
enum Command {
    #[command(arg_required_else_help = true)]
    Serve(ServeCommand),
    #[command()]
    Configuration(ConfigurationCommand),
    #[command()]
    CheckHealth(CheckHealthCommand),
}

#[derive(Clone, Parser)]
struct ServeCommand {
    #[arg(long, value_name = "CONFIGURATION_FILE", env = "CONFIGURATION_FILE")]
    configuration: String,
    #[arg(long, value_name = "OTLP_ENDPOINT", env = "OTLP_ENDPOINT")]
    otlp_endpoint: Option<String>, // NOTE: `tracing` crate uses `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` ENV variable, but we want to control the endpoint via CLI interface
    #[arg(long, value_name = "PORT", env = "PORT", default_value = "8100")]
    port: Port,
    #[arg(long, value_name = "SERVICE_TOKEN_SECRET", env = "SERVICE_TOKEN_SECRET")]
    service_token_secret: Option<String>,
}

#[derive(Clone, Parser)]
struct ConfigurationCommand {
    #[command(subcommand)]
    command: ConfigurationSubcommand,
}

#[derive(Clone, Subcommand)]
enum ConfigurationSubcommand {
    #[command()]
    Serve(ServeConfigurationCommand),
}

#[derive(Clone, Parser)]
struct ServeConfigurationCommand {
    #[arg(long, value_name = "PORT", env = "PORT", default_value = "9100")]
    port: Port,
}

#[derive(Clone, Parser)]
struct CheckHealthCommand {
    #[arg(long, value_name = "HOST")]
    host: Option<String>,
    #[arg(long, value_name = "PORT", env = "PORT", default_value = "8100")]
    port: Port,
}

type Port = u16;

#[derive(Debug, Clone)]
pub struct ServerState<C: Connector> {
    configuration: C::Configuration,
    state: C::State,
    metrics: Registry,
}

/// A default main function for a connector.
///
/// The intent is that this function can replace your `main` function
/// entirely, if you implement the [`Connector`] trait:
///
/// ```ignore
/// struct MyConnector { /* ... */ }
///
/// impl Connector for MyConnector { /* ... */ }
///
/// #[tokio::main]
/// async fn main() {
///     default_main::<MyConnector>().await.unwrap()
/// }
/// ```
///
/// This function adopts certain conventions for features which are
/// not described in the [NDC specification](http://hasura.github.io/ndc-spec/).
/// Specifically:
///
/// - It reads configuration as JSON from a file specified on the command line,
/// - It reports traces to an OTLP collector specified on the command line,
/// - Logs are written to stdout
pub async fn default_main<C: Connector + Clone + Default + 'static>() -> Result<(), Box<dyn Error>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Sync + Send + Clone,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    let CliArgs { command } = CliArgs::parse();

    match command {
        Command::Serve(serve_command) => serve::<C>(serve_command).await,
        Command::Configuration(configure_command) => configuration::<C>(configure_command).await,
        Command::CheckHealth(check_health_command) => check_health(check_health_command).await,
    }
}

fn init_tracing(serve_command: &ServeCommand) -> Result<(), Box<dyn Error>> {
    global::set_text_map_propagator(TraceContextPropagator::new());

    let tracer = opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(
            opentelemetry_otlp::new_exporter().tonic().with_endpoint(
                serve_command
                    .otlp_endpoint
                    .clone()
                    .unwrap_or(OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT.into()),
            ),
        )
        .with_trace_config(
            opentelemetry::sdk::trace::config()
                .with_resource(opentelemetry::sdk::Resource::new(vec![
                    KeyValue::new(
                        opentelemetry_semantic_conventions::resource::SERVICE_NAME,
                        "ndc-hub",
                    ),
                    KeyValue::new(
                        opentelemetry_semantic_conventions::resource::SERVICE_VERSION,
                        env!("CARGO_PKG_VERSION"),
                    ),
                ]))
                .with_sampler(Sampler::ParentBased(Box::new(Sampler::AlwaysOn))),
        )
        .install_batch(opentelemetry::runtime::Tokio)?;

    tracing_subscriber::registry()
        .with(
            tracing_opentelemetry::layer()
                .with_exception_field_propagation(true)
                .with_tracer(tracer),
        )
        .with(EnvFilter::builder().parse("info,otel::tracing=trace,otel=debug")?)
        .with(
            tracing_subscriber::fmt::layer()
                .json()
                .with_timer(tracing_subscriber::fmt::time::time()),
        )
        .init();

    Ok(())
}

async fn serve<C: Connector + Clone + Default + 'static>(
    serve_command: ServeCommand,
) -> Result<(), Box<dyn Error>>
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    init_tracing(&serve_command).expect("Unable to initialize tracing");

    let server_state = init_server_state::<C>(serve_command.configuration).await;

    let expected_auth_header: Option<HeaderValue> = serve_command.service_token_secret.and_then(|service_token_secret| {
        let expected_bearer = format!("Bearer {}", service_token_secret); // TODO
        HeaderValue::from_str(&expected_bearer).ok()
    }); 

    let router = create_router::<C>(server_state)
        .layer(
            TraceLayer::new_for_http().make_span_with(DefaultMakeSpan::default().level(Level::INFO)),
        ).layer(
            ValidateRequestHeaderLayer::custom(move |request: &mut Request<Body>| {
                // Validate the request
                let auth_header = request.headers().get("Authorization")
                    .map(|v| v.clone());

                // NOTE: The comparison should probably be more permissive to allow for whitespace, etc.
                if auth_header == expected_auth_header { return Ok(()); }
                Err((StatusCode::UNAUTHORIZED, "").into_response())
            })
        );

    let port = serve_command.port;
    let address = net::SocketAddr::new(net::IpAddr::V4(net::Ipv4Addr::UNSPECIFIED), port);

    println!("Starting server on {}", address);

    axum::Server::bind(&address)
        .serve(router.into_make_service())
        .with_graceful_shutdown(async {
            tokio::signal::ctrl_c()
                .await
                .expect("unable to install signal handler");

            opentelemetry::global::shutdown_tracer_provider();
        })
        .await?;

    Ok(())
}

/// Initialize the server state from the configuration file.
pub async fn init_server_state<C: Connector + Clone + Default + 'static>(
    config_file: String,
) -> ServerState<C>
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    let configuration_json = std::fs::read_to_string(config_file).unwrap();
    let raw_configuration =
        serde_json::de::from_str::<C::RawConfiguration>(configuration_json.as_str()).unwrap();
    let configuration = C::validate_raw_configuration(&raw_configuration)
        .await
        .unwrap();

    let mut metrics = Registry::new();
    let state = C::try_init_state(&configuration, &mut metrics)
        .await
        .unwrap();

    ServerState::<C> {
        configuration,
        state,
        metrics,
    }
}

pub fn create_router<C: Connector + Clone + 'static>(state: ServerState<C>) -> Router
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + Clone + Sync + Send,
    C::State: Sync + Send + Clone,
{
    Router::new()
        .route("/capabilities", get(get_capabilities::<C>))
        .route("/health", get(get_health::<C>))
        .route("/metrics", get(get_metrics::<C>))
        .route("/schema", get(get_schema::<C>))
        .route("/query", post(post_query::<C>))
        .route("/explain", post(post_explain::<C>))
        .route("/mutation", post(post_mutation::<C>))
        .with_state(state)
}

async fn get_metrics<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<String, (StatusCode, Json<ErrorResponse>)> {
    routes::get_metrics::<C>(&state.configuration, &state.state, state.metrics)
}

async fn get_capabilities<C: Connector>() -> Json<CapabilitiesResponse> {
    routes::get_capabilities::<C>().await
}

async fn get_health<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<(), (StatusCode, Json<ErrorResponse>)> {
    routes::get_health::<C>(&state.configuration, &state.state).await
}

async fn get_schema<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<Json<SchemaResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::get_schema::<C>(&state.configuration).await
}

async fn post_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<QueryRequest>,
) -> Result<Json<ExplainResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_explain::<C>(&state.configuration, &state.state, request).await
}

async fn post_mutation<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<MutationRequest>,
) -> Result<Json<MutationResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_mutation::<C>(&state.configuration, &state.state, request).await
}

async fn post_query<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<QueryRequest>,
) -> Result<Json<QueryResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_query::<C>(&state.configuration, &state.state, request).await
}

async fn configuration<C: Connector + 'static>(
    command: ConfigurationCommand,
) -> Result<(), Box<dyn Error>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Clone + Sync + Send,
    C::Configuration: Sync + Send,
{
    match command.command {
        ConfigurationSubcommand::Serve(serve_command) => {
            serve_configuration::<C>(serve_command).await
        }
    }
}

async fn serve_configuration<C: Connector + 'static>(
    serve_command: ServeConfigurationCommand,
) -> Result<(), Box<dyn Error>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Sync + Send,
    C::Configuration: Sync + Send,
{
    let port = serve_command.port;
    let address = net::SocketAddr::new(net::IpAddr::V4(net::Ipv4Addr::UNSPECIFIED), port);

    println!("Starting server on {}", address);

    let cors = CorsLayer::new()
        .allow_origin(tower_http::cors::Any)
        .allow_headers(tower_http::cors::Any);

    let router = Router::new()
        .route("/", get(get_empty::<C>).post(post_update::<C>))
        .route("/schema", get(get_config_schema::<C>))
        .route("/validate", post(post_validate::<C>))
        .layer(cors);

    axum::Server::bind(&address)
        .serve(router.into_make_service())
        .with_graceful_shutdown(async {
            tokio::signal::ctrl_c()
                .await
                .expect("unable to install signal handler");
        })
        .await?;

    Ok(())
}

async fn get_empty<C: Connector>() -> Json<C::RawConfiguration>
where
    C::RawConfiguration: Serialize,
{
    Json(C::make_empty_configuration())
}

async fn post_update<C: Connector>(
    Json(configuration): Json<C::RawConfiguration>,
) -> Result<Json<C::RawConfiguration>, (StatusCode, String)>
where
    C::RawConfiguration: Serialize + DeserializeOwned,
{
    let updated = C::update_configuration(&configuration)
        .await
        .map_err(|err| match err {
            UpdateConfigurationError::Other(err) => {
                (StatusCode::INTERNAL_SERVER_ERROR, err.to_string())
            }
        })?;
    Ok(Json(updated))
}

async fn get_config_schema<C: Connector>() -> Json<RootSchema>
where
    C::RawConfiguration: JsonSchema,
{
    let schema = schemars::schema_for!(C::RawConfiguration);
    Json(schema)
}

#[derive(Debug, Clone, Serialize)]
struct ValidateResponse {
    schema: SchemaResponse,
}

#[derive(Debug, Clone, Serialize)]
#[serde(tag = "type")]
enum ValidateErrors {
    InvalidConfiguration { ranges: Vec<InvalidRange> },
    UnableToBuildSchema,
}

async fn post_validate<C: Connector>(
    Json(configuration): Json<C::RawConfiguration>,
) -> Result<Json<ValidateResponse>, (StatusCode, Json<ValidateErrors>)>
where
    C::RawConfiguration: DeserializeOwned,
{
    let configuration = C::validate_raw_configuration(&configuration)
        .await
        .map_err(|e| match e {
            crate::connector::ValidateError::ValidateError(ranges) => (
                StatusCode::BAD_REQUEST,
                Json(ValidateErrors::InvalidConfiguration { ranges }),
            ),
        })?;
    let schema = C::get_schema(&configuration).await.map_err(|e| match e {
        SchemaError::Other(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ValidateErrors::UnableToBuildSchema),
        ),
    })?;
    Ok(Json(ValidateResponse { schema }))
}

async fn check_health(
    CheckHealthCommand { host, port }: CheckHealthCommand,
) -> Result<(), Box<dyn Error>> {
    match check_health::check_health(host, port).await {
        Ok(()) => {
            println!("Health check succeeded.");
            Ok(())
        }
        Err(err) => {
            println!("Health check failed.");
            Err(err.into())
        }
    }
}

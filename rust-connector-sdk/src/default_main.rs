mod v2_compat;

use crate::{
    check_health,
    connector::{Connector, InvalidRange, SchemaError, UpdateConfigurationError},
    json_rejection::JsonRejection,
    json_response::JsonResponse,
    routes,
    tracing::{init_tracing, make_span, on_response},
};
use axum_extra::extract::WithRejection;

use std::error::Error;
use std::net;
use std::path::{Path, PathBuf};
use std::process::exit;

use async_trait::async_trait;
use axum::{
    body::Body,
    extract::State,
    http::{HeaderValue, Request, StatusCode},
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use base64::{engine::general_purpose, Engine};
use clap::{Parser, Subcommand};
use ndc_client::models::{
    CapabilitiesResponse, ErrorResponse, ExplainResponse, MutationRequest, MutationResponse,
    QueryRequest, QueryResponse, SchemaResponse,
};
use ndc_test::report;
use prometheus::Registry;
use schemars::{schema::RootSchema, JsonSchema};
use serde::{de::DeserializeOwned, Serialize};
use tower_http::{
    cors::CorsLayer, trace::TraceLayer, validate_request::ValidateRequestHeaderLayer,
};

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
    Test(TestCommand),
    #[command()]
    Replay(ReplayCommand),
    #[command()]
    CheckHealth(CheckHealthCommand),
}

#[derive(Clone, Parser)]
struct ServeCommand {
    #[arg(long, value_name = "CONFIGURATION_FILE", env = "CONFIGURATION_FILE")]
    configuration: PathBuf,
    #[arg(long, value_name = "OTLP_ENDPOINT", env = "OTLP_ENDPOINT")]
    otlp_endpoint: Option<String>, // NOTE: `tracing` crate uses `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` ENV variable, but we want to control the endpoint via CLI interface
    #[arg(long, value_name = "PORT", env = "PORT", default_value = "8100")]
    port: Port,
    #[arg(
        long,
        value_name = "SERVICE_TOKEN_SECRET",
        env = "SERVICE_TOKEN_SECRET"
    )]
    service_token_secret: Option<String>,
    #[arg(long, value_name = "OTEL_SERVICE_NAME", env = "OTEL_SERVICE_NAME")]
    service_name: Option<String>,
    #[arg(long, env = "ENABLE_V2_COMPATIBILITY")]
    enable_v2_compatibility: bool,
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
    #[arg(long, value_name = "OTEL_SERVICE_NAME", env = "OTEL_SERVICE_NAME")]
    service_name: Option<String>,
    #[arg(long, value_name = "OTLP_ENDPOINT", env = "OTLP_ENDPOINT")]
    otlp_endpoint: Option<String>, // NOTE: `tracing` crate uses `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` ENV variable, but we want to control the endpoint via CLI interface
}

#[derive(Clone, Parser)]
struct TestCommand {
    #[arg(long, value_name = "SEED", env = "SEED")]
    seed: Option<String>,
    #[arg(long, value_name = "CONFIGURATION_FILE", env = "CONFIGURATION_FILE")]
    configuration: PathBuf,
    #[arg(long, value_name = "DIRECTORY", env = "SNAPSHOTS_DIR")]
    snapshots_dir: Option<PathBuf>,
}

#[derive(Clone, Parser)]
struct ReplayCommand {
    #[arg(long, value_name = "CONFIGURATION_FILE", env = "CONFIGURATION_FILE")]
    configuration: PathBuf,
    #[arg(long, value_name = "DIRECTORY", env = "SNAPSHOTS_DIR")]
    snapshots_dir: PathBuf,
}

#[derive(Clone, Parser)]
struct CheckHealthCommand {
    #[arg(long, value_name = "HOST")]
    host: Option<String>,
    #[arg(long, value_name = "PORT", env = "PORT", default_value = "8100")]
    port: Port,
}

type Port = u16;

#[derive(Debug)]
pub struct ServerState<C: Connector> {
    configuration: C::Configuration,
    state: C::State,
    metrics: Registry,
}

impl<C: Connector> Clone for ServerState<C>
where
    C::Configuration: Clone,
    C::State: Clone,
{
    fn clone(&self) -> Self {
        Self {
            configuration: self.configuration.clone(),
            state: self.state.clone(),
            metrics: self.metrics.clone(),
        }
    }
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
pub async fn default_main<C: Connector + Default + 'static>(
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    let CliArgs { command } = CliArgs::parse();

    match command {
        Command::Serve(serve_command) => serve::<C>(serve_command).await,
        Command::Configuration(configure_command) => configuration::<C>(configure_command).await,
        Command::Test(test_command) => test::<C>(test_command).await,
        Command::Replay(replay_command) => replay::<C>(replay_command).await,
        Command::CheckHealth(check_health_command) => check_health(check_health_command).await,
    }
}

async fn serve<C: Connector + Default + 'static>(
    serve_command: ServeCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    init_tracing(&serve_command.service_name, &serve_command.otlp_endpoint)
        .expect("Unable to initialize tracing");

    let server_state = init_server_state::<C>(serve_command.configuration).await;

    let router = create_router::<C>(
        server_state.clone(),
        serve_command.service_token_secret.clone(),
    );

    let router = if serve_command.enable_v2_compatibility {
        let v2_router = create_v2_router(server_state, serve_command.service_token_secret.clone());
        Router::new().merge(router).nest("/v2", v2_router)
    } else {
        router
    };

    let port = serve_command.port;
    let address = net::SocketAddr::new(net::IpAddr::V4(net::Ipv4Addr::UNSPECIFIED), port);

    println!("Starting server on {}", address);

    axum::Server::bind(&address)
        .serve(router.into_make_service())
        .with_graceful_shutdown(async {
            // wait for a SIGINT, i.e. a Ctrl+C from the keyboard
            let sigint = async {
                tokio::signal::ctrl_c()
                    .await
                    .expect("unable to install signal handler")
            };
            // wait for a SIGTERM, i.e. a normal `kill` command
            #[cfg(unix)]
            let sigterm = async {
                tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
                    .expect("failed to install signal handler")
                    .recv()
                    .await
            };
            // block until either of the above happens
            #[cfg(unix)]
            tokio::select! {
                _ = sigint => (),
                _ = sigterm => (),
            }
            #[cfg(windows)]
            tokio::select! {
                _ = sigint => (),
            }

            opentelemetry::global::shutdown_tracer_provider();
        })
        .await?;

    Ok(())
}

/// Initialize the server state from the configuration file.
pub async fn init_server_state<C: Connector + Default + 'static>(
    config_file: impl AsRef<Path>,
) -> ServerState<C>
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    let configuration_json = std::fs::read_to_string(config_file).unwrap();
    let raw_configuration =
        serde_json::de::from_str::<C::RawConfiguration>(configuration_json.as_str()).unwrap();
    let configuration = C::validate_raw_configuration(raw_configuration)
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

pub fn create_router<C: Connector + 'static>(
    state: ServerState<C>,
    service_token_secret: Option<String>,
) -> Router
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + Clone + Sync + Send,
    C::State: Sync + Send + Clone,
{
    let router = Router::new()
        .route("/capabilities", get(get_capabilities::<C>))
        .route("/health", get(get_health::<C>))
        .route("/metrics", get(get_metrics::<C>))
        .route("/schema", get(get_schema::<C>))
        .route("/query", post(post_query::<C>))
        .route("/explain", post(post_explain::<C>))
        .route("/mutation", post(post_mutation::<C>))
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(make_span)
                .on_response(on_response)
                .on_failure(|err, _dur, _span: &tracing::Span| {
                    tracing::error!(
                        meta.signal_type = "log",
                        event.domain = "ndc",
                        event.name = "Request failure",
                        name = "Request failure",
                        body = %err,
                        error = true,
                    )
                }),
        )
        .with_state(state);

    let expected_auth_header: Option<HeaderValue> =
        service_token_secret.and_then(|service_token_secret| {
            let expected_bearer = format!("Bearer {}", service_token_secret);
            HeaderValue::from_str(&expected_bearer).ok()
        });

    router
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(make_span)
                .on_response(on_response),
        )
        .layer(ValidateRequestHeaderLayer::custom(
            move |request: &mut Request<Body>| {
                // Validate the request
                let auth_header = request.headers().get("Authorization").cloned();

                // NOTE: The comparison should probably be more permissive to allow for whitespace, etc.
                if auth_header == expected_auth_header {
                    return Ok(());
                }

                let message = "Bearer token does not match.".to_string();

                tracing::error!(
                    meta.signal_type = "log",
                    event.domain = "ndc",
                    event.name = "Authorization error",
                    name = "Authorization error",
                    body = message,
                    error = true,
                );
                Err((
                    StatusCode::UNAUTHORIZED,
                    Json(ErrorResponse {
                        message: "Internal error".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "cause".into(),
                            serde_json::Value::String(message),
                        )])),
                    }),
                )
                    .into_response())
            },
        ))
}

pub fn create_v2_router<C: Connector + 'static>(
    state: ServerState<C>,
    service_token_secret: Option<String>,
) -> Router
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + Clone + Sync + Send,
    C::State: Sync + Send + Clone,
{
    Router::new()
        .route("/schema", post(v2_compat::post_schema::<C>))
        .route("/query", post(v2_compat::post_query::<C>))
        // .route("/mutation", post(v2_compat::post_mutation::<C>))
        // .route("/raw", post(v2_compat::post_raw::<C>))
        .route("/explain", post(v2_compat::post_explain::<C>))
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(make_span)
                .on_response(on_response),
        )
        .layer(ValidateRequestHeaderLayer::custom(
            move |request: &mut Request<Body>| {
                let provided_service_token_secret = request
                    .headers()
                    .get("x-hasura-dataconnector-config")
                    .and_then(|config_header| {
                        serde_json::from_slice::<v2_compat::SourceConfig>(config_header.as_bytes())
                            .ok()
                    })
                    .and_then(|config| config.service_token_secret);

                if service_token_secret == provided_service_token_secret {
                    // if token set & config header present & values match
                    // or token not set & config header not set/does not have value for token key
                    // allow request
                    Ok(())
                } else {
                    // all other cases, block request
                    let message = "Service Token Secret does not match.".to_string();

                    tracing::error!(
                        meta.signal_type = "log",
                        event.domain = "ndc",
                        event.name = "Authorization error",
                        name = "Authorization error",
                        body = message,
                        error = true,
                    );
                    Err((
                        StatusCode::UNAUTHORIZED,
                        Json(ErrorResponse {
                            message: "Internal error".into(),
                            details: serde_json::Value::Object(serde_json::Map::from_iter([(
                                "cause".into(),
                                serde_json::Value::String(message),
                            )])),
                        }),
                    )
                        .into_response())
                }
            },
        ))
        // capabilities and health endpoints are exempt from auth requirements
        .route("/capabilities", get(v2_compat::get_capabilities::<C>))
        .route("/health", get(v2_compat::get_health))
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(make_span)
                .on_response(on_response)
                .on_failure(|err, _dur, _span: &_| {
                    tracing::error!(
                        meta.signal_type = "log",
                        event.domain = "ndc",
                        event.name = "Request failure",
                        name = "Request failure",
                        body = %err,
                        error = true,
                    );
                }),
        )
        .with_state(state)
}

async fn get_metrics<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<String, (StatusCode, Json<ErrorResponse>)> {
    routes::get_metrics::<C>(&state.configuration, &state.state, state.metrics)
}

async fn get_capabilities<C: Connector>() -> JsonResponse<CapabilitiesResponse> {
    routes::get_capabilities::<C>().await
}

async fn get_health<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<(), (StatusCode, Json<ErrorResponse>)> {
    routes::get_health::<C>(&state.configuration, &state.state).await
}

async fn get_schema<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<JsonResponse<SchemaResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::get_schema::<C>(&state.configuration).await
}

async fn post_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    WithRejection(Json(request), _): WithRejection<Json<QueryRequest>, JsonRejection>,
) -> Result<JsonResponse<ExplainResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_explain::<C>(&state.configuration, &state.state, request).await
}

async fn post_mutation<C: Connector>(
    State(state): State<ServerState<C>>,
    WithRejection(Json(request), _): WithRejection<Json<MutationRequest>, JsonRejection>,
) -> Result<JsonResponse<MutationResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_mutation::<C>(&state.configuration, &state.state, request).await
}

async fn post_query<C: Connector>(
    State(state): State<ServerState<C>>,
    WithRejection(Json(request), _): WithRejection<Json<QueryRequest>, JsonRejection>,
) -> Result<JsonResponse<QueryResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_query::<C>(&state.configuration, &state.state, request).await
}

async fn configuration<C: Connector + 'static>(
    command: ConfigurationCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Sync + Send,
    C::Configuration: Sync + Send + Serialize,
{
    match command.command {
        ConfigurationSubcommand::Serve(serve_command) => {
            serve_configuration::<C>(serve_command).await
        }
    }
}

async fn serve_configuration<C: Connector + 'static>(
    serve_command: ServeConfigurationCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: Serialize + DeserializeOwned + JsonSchema + Sync + Send,
    C::Configuration: Sync + Send + Serialize,
{
    let port = serve_command.port;
    let address = net::SocketAddr::new(net::IpAddr::V4(net::Ipv4Addr::UNSPECIFIED), port);

    init_tracing(&serve_command.service_name, &serve_command.otlp_endpoint)
        .expect("Unable to initialize tracing");

    println!("Starting server on {}", address);

    let cors = CorsLayer::new()
        .allow_origin(tower_http::cors::Any)
        .allow_headers(tower_http::cors::Any);

    let router = Router::new()
        .route("/", get(get_empty::<C>).post(post_update::<C>))
        .route("/schema", get(get_config_schema::<C>))
        .route("/validate", post(post_validate::<C>))
        .route("/health", get(|| async {}))
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(make_span)
                .on_response(on_response)
                .on_failure(|err, _dur, _span: &_| {
                    tracing::error!(
                        meta.signal_type = "log",
                        event.domain = "ndc",
                        event.name = "Request failure",
                        name = "Request failure",
                        body = %err,
                        error = true,
                    );
                }),
        )
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
    WithRejection(Json(configuration), _): WithRejection<Json<C::RawConfiguration>, JsonRejection>,
) -> Result<Json<C::RawConfiguration>, (StatusCode, String)>
where
    C::RawConfiguration: Serialize + DeserializeOwned,
{
    let updated = C::update_configuration(configuration)
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
    capabilities: CapabilitiesResponse,
    resolved_configuration: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(tag = "type")]
enum ValidateErrors {
    InvalidConfiguration { ranges: Vec<InvalidRange> },
    UnableToBuildSchema,
    UnableToBuildCapabilities,
    JsonEncodingError(String),
}

async fn post_validate<C: Connector>(
    WithRejection(Json(configuration), _): WithRejection<Json<C::RawConfiguration>, JsonRejection>,
) -> Result<Json<ValidateResponse>, (StatusCode, Json<ValidateErrors>)>
where
    C::RawConfiguration: DeserializeOwned,
    C::Configuration: Serialize,
{
    let configuration =
        C::validate_raw_configuration(configuration)
            .await
            .map_err(|e| match e {
                crate::connector::ValidateError::ValidateError(ranges) => (
                    StatusCode::BAD_REQUEST,
                    Json(ValidateErrors::InvalidConfiguration { ranges }),
                ),
            })?;
    let schema = C::get_schema(&configuration)
        .await
        .and_then(JsonResponse::into_value)
        .map_err(|e| match e {
            SchemaError::Other(err) => {
                tracing::error!(
                    meta.signal_type = "log",
                    event.domain = "ndc",
                    event.name = "Unable to build schema",
                    name = "Unable to build schema",
                    body = %err,
                    error = true,
                );
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ValidateErrors::UnableToBuildSchema),
                )
            }
        })?;
    let capabilities =
        C::get_capabilities()
            .await
            .into_value()
            .map_err(|e: Box<dyn Error + Send + Sync>| {
                tracing::error!(
                    meta.signal_type = "log",
                    event.domain = "ndc",
                    event.name = "Unable to build capabilities",
                    name = "Unable to build capabilities",
                    body = %e,
                    error = true,
                );
                (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ValidateErrors::UnableToBuildCapabilities),
                )
            })?;
    let resolved_config_bytes = serde_json::to_vec(&configuration).map_err(|err| {
        tracing::error!(
            meta.signal_type = "log",
            event.domain = "ndc",
            event.name = "Unable to serialize validated configuration",
            name = "Unable to serialize validated configuration",
            body = %err,
            error = true,
        );
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ValidateErrors::JsonEncodingError(err.to_string())),
        )
    })?;
    let resolved_configuration = general_purpose::STANDARD.encode(resolved_config_bytes);
    Ok(Json(ValidateResponse {
        schema,
        capabilities,
        resolved_configuration,
    }))
}

struct ConnectorAdapter<C: Connector> {
    configuration: C::Configuration,
    state: C::State,
}

#[async_trait]
impl<C: Connector> ndc_test::Connector for ConnectorAdapter<C>
where
    C::Configuration: Send + Sync + 'static,
    C::State: Send + Sync + 'static,
{
    async fn get_capabilities(
        &self,
    ) -> Result<ndc_client::models::CapabilitiesResponse, ndc_test::Error> {
        C::get_capabilities()
            .await
            .into_value::<Box<dyn std::error::Error + Send + Sync>>()
            .map_err(|err| ndc_test::Error::OtherError(err))
    }

    async fn get_schema(&self) -> Result<ndc_client::models::SchemaResponse, ndc_test::Error> {
        match C::get_schema(&self.configuration).await {
            Ok(response) => response
                .into_value::<Box<dyn std::error::Error + Send + Sync>>()
                .map_err(|err| ndc_test::Error::OtherError(err)),
            Err(err) => Err(ndc_test::Error::OtherError(err.into())),
        }
    }

    async fn query(
        &self,
        request: ndc_client::models::QueryRequest,
    ) -> Result<ndc_client::models::QueryResponse, ndc_test::Error> {
        match C::query(&self.configuration, &self.state, request)
            .await
            .and_then(JsonResponse::into_value)
        {
            Ok(response) => Ok(response),
            Err(err) => Err(ndc_test::Error::OtherError(err.into())),
        }
    }

    async fn mutation(
        &self,
        request: ndc_client::models::MutationRequest,
    ) -> Result<ndc_client::models::MutationResponse, ndc_test::Error> {
        match C::mutation(&self.configuration, &self.state, request)
            .await
            .and_then(JsonResponse::into_value)
        {
            Ok(response) => Ok(response),
            Err(err) => Err(ndc_test::Error::OtherError(err.into())),
        }
    }
}

async fn test<C: Connector + 'static>(
    command: TestCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: DeserializeOwned,
    C::Configuration: Sync + Send + 'static,
    C::State: Send + Sync + 'static,
{
    let test_configuration = ndc_test::TestConfiguration {
        seed: command.seed,
        snapshots_dir: command.snapshots_dir,
    };

    let connector = make_connector_adapter::<C>(command.configuration).await;
    let results = ndc_test::test_connector(&test_configuration, &connector).await;

    if !results.failures.is_empty() {
        println!();
        println!("{}", report(results));

        exit(1)
    }

    Ok(())
}

async fn replay<C: Connector + 'static>(
    command: ReplayCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::RawConfiguration: DeserializeOwned,
    C::Configuration: Sync + Send + 'static,
    C::State: Send + Sync + 'static,
{
    let connector = make_connector_adapter::<C>(command.configuration).await;
    let results = ndc_test::test_snapshots_in_directory(&connector, command.snapshots_dir).await;

    if !results.failures.is_empty() {
        println!();
        println!("{}", report(results));

        exit(1)
    }

    Ok(())
}

async fn make_connector_adapter<C: Connector + 'static>(
    configuration_path: impl AsRef<Path>,
) -> ConnectorAdapter<C>
where
    C::RawConfiguration: DeserializeOwned,
{
    let configuration_json = std::fs::read_to_string(configuration_path).unwrap();
    let raw_configuration =
        serde_json::de::from_str::<C::RawConfiguration>(configuration_json.as_str()).unwrap();
    let configuration = C::validate_raw_configuration(raw_configuration)
        .await
        .unwrap();

    let mut metrics = Registry::new();
    let state = C::try_init_state(&configuration, &mut metrics)
        .await
        .unwrap();

    ConnectorAdapter::<C> {
        configuration,
        state,
    }
}

async fn check_health(
    CheckHealthCommand { host, port }: CheckHealthCommand,
) -> Result<(), Box<dyn Error + Send + Sync>> {
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

use crate::{connector::Connector, routes};
use axum::{
    extract::State,
    http::StatusCode,
    routing::{get, post},
    Json, Router
};
use clap::{Args, Parser, Subcommand};
use ndc_client::models::{
    CapabilitiesResponse, ExplainResponse, MutationRequest, MutationResponse, QueryRequest,
    QueryResponse, SchemaResponse,
};

use prometheus::Registry;
use serde::{de::DeserializeOwned, Serialize};
use std::{error::Error, env};
use std::net::SocketAddr;
use tower_http::trace::TraceLayer;
use axum_tracing_opentelemetry::{opentelemetry_tracing_layer, response_with_trace_layer};

#[derive(Parser)]
struct CliArgs<C: Connector>
where
    C::ConfigureArgs: Clone + Send + Sync + Args,
{
    #[command(subcommand)]
    command: Command<C>,
}

#[derive(Clone, Subcommand)]
enum Command<C: Connector>
where
    C::ConfigureArgs: Clone + Send + Sync + Args,
{
    #[command(arg_required_else_help = true)]
    Serve(ServeCommand),
    #[command()]
    GenerateConfiguration(C::ConfigureArgs),
}

#[derive(Clone, Parser)]
struct ServeCommand {
    #[arg(long, value_name = "CONFIGURATION_FILE", env = "CONFIGURATION_FILE")]
    configuration: String,
    #[arg(long, value_name = "OTLP_ENDPOINT", env = "OTLP_ENDPOINT")]
    otlp_endpoint: Option<String>, // NOTE: `tracing` crate uses `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` ENV variable, but we want to control the endpoint via CLI interface
    #[arg(long, value_name = "PORT", env = "PORT")]
    port: Option<String>,
}

#[derive(Debug, Clone)]
pub struct ServerState<C: Connector> {
    configuration: C::Configuration,
    state: C::State,
    metrics: Registry,
    // tracer: Arc<sdk::trace::Tracer>,
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
    C::ConfigureArgs: Clone + Send + Sync + Args,
    C::RawConfiguration: Serialize + DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    let CliArgs { command } = CliArgs::<C>::parse();

    match command {
        Command::Serve(serve_command) => serve::<C>(serve_command).await,
        Command::GenerateConfiguration(configure_command) => {
            configure::<C>(configure_command).await
        }
    }
}

async fn serve<C: Connector + Clone + Default + 'static>(
    serve_command: ServeCommand,
) -> Result<(), Box<dyn Error>>
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + DeserializeOwned + Sync + Send + Clone,
    C::State: Sync + Send + Clone,
{
    // Set endpoint ENV picked up by macros in `traces` crate via CLI option if used
    // TODO: Check that tracing library doesn't have a better way to do this.
    serve_command.otlp_endpoint.map(|e| {
        env::set_var(opentelemetry_otlp::OTEL_EXPORTER_OTLP_TRACES_ENDPOINT, e);
    });

    axum_tracing_opentelemetry::tracing_subscriber_ext::init_subscribers().expect("Couldn't init OT");

    let server_state = init_server_state::<C>(serve_command.configuration).await;

    let router = create_router::<C>(server_state);

    let port = serve_command.port.unwrap_or("8100".into());
    let address = SocketAddr::new("0.0.0.0".parse()?, port.parse()?);

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

pub fn create_router<'a, C: Connector + Clone + 'static>(state: ServerState<C>) -> Router
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
        .layer(TraceLayer::new_for_http())
        .layer(response_with_trace_layer())
        .layer(opentelemetry_tracing_layer())
}

async fn get_metrics<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<String, StatusCode> {
    routes::get_metrics::<C>(state.metrics)
}

async fn get_capabilities<C: Connector>() -> Json<CapabilitiesResponse> {
    routes::get_capabilities::<C>().await
}

async fn get_health<C: Connector>(State(state): State<ServerState<C>>) -> StatusCode {
    routes::get_health::<C>(&state.configuration, &state.state).await
}

async fn get_schema<C: Connector>(
    State(state): State<ServerState<C>>,
) -> Result<Json<SchemaResponse>, StatusCode> {
    routes::get_schema::<C>(&state.configuration).await
}

async fn post_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<QueryRequest>,
) -> Result<Json<ExplainResponse>, StatusCode> {
    routes::post_explain::<C>(&state.configuration, &state.state, request).await
}

async fn post_mutation<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<MutationRequest>,
) -> Result<Json<MutationResponse>, StatusCode> {
    routes::post_mutation::<C>(&state.configuration, &state.state, request).await
}

async fn post_query<C: Connector>(
    State(state): State<ServerState<C>>,
    request: Json<QueryRequest>,
) -> Result<Json<QueryResponse>, StatusCode> {
    routes::post_query::<C>(&state.configuration, &state.state, request).await
}

async fn configure<C: Connector + Clone + Default + 'static>(
    args: C::ConfigureArgs,
) -> Result<(), Box<dyn Error>>
where
    C::ConfigureArgs: Clone + Send + Sync + Args,
    C::RawConfiguration: Serialize,
{
    let configuration = C::configure(&args).await?;
    println!("{}", serde_json::to_string_pretty(&configuration)?);
    Ok(())
}

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
use opentelemetry_otlp::{OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT};
// use opentelemetry_sdk::{propagation::TraceContextPropagator, trace::Tracer};
use prometheus::Registry;
use serde::{de::DeserializeOwned, Serialize};
use std::error::Error;
use std::net::SocketAddr;
use tower_http::trace::TraceLayer;
// use opentelemetry::trace::Tracer as Tracer2;
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
    otlp_endpoint: Option<String>,
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

    axum_tracing_opentelemetry::tracing_subscriber_ext::init_subscribers().expect("Couldn't init OT");

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
    // global::set_text_map_propagator(TraceContextPropagator::new());

    // tracing_subscriber::fmt()
    //     .with_max_level(tracing::Level::DEBUG)
    //     .init();

    let server_state = init_server_state::<C>(serve_command.configuration /*  , serve_command.otlp_endpoint */).await; // Shouldn't this be initialized prior to request handling?

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
    // otlp_endpoint: Option<String>,
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

    // let tracer = opentelemetry_otlp::new_pipeline()
    //     .tracing()
    //     .with_exporter(
    //         opentelemetry_otlp::new_exporter().tonic().with_endpoint(
    //             otlp_endpoint
    //                 .unwrap_or(OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT.into()),
    //         ),
    //     )
    //     .with_trace_config(opentelemetry::sdk::trace::config().with_resource(
    //         opentelemetry_sdk::Resource::new(vec![
    //             KeyValue::new(
    //                 opentelemetry_semantic_conventions::resource::SERVICE_NAME,
    //                 "ndc-hub",
    //             ),
    //             KeyValue::new(
    //                 opentelemetry_semantic_conventions::resource::SERVICE_VERSION,
    //                 env!("CARGO_PKG_VERSION"),
    //             ),
    //         ]),
    //     ))
    //     .install_batch(opentelemetry::runtime::Tokio)
    //     .expect("failed to create tracer");

    ServerState::<C> {
        configuration,
        state,
        metrics,
        // tracer:Arc::new(tracer),
    }
}

// // Example taken from https://docs.rs/axum/latest/axum/middleware/fn.from_fn_with_state.html
// async fn otlp_middleware<B>(
//     State(tracer): State<Arc<sdk::trace::Tracer>>,
//     // you can add more extractors here but the last
//     // extractor must implement `FromRequest` which
//     // `Request` does
//     request: Request<B>,
//     next: Next<B>,
// ) -> axum::response::Response {
//     let response: Response<BoxBody> = tracer.in_span("request", |cx| async {
//         // TODO: Enrich with MD about route, etc.
//         next.run(request).await
//     }).await;

//     response
// }

pub fn create_router<'a, C: Connector + Clone + 'static>(state: ServerState<C>) -> Router
where
    C::RawConfiguration: DeserializeOwned + Sync + Send,
    C::Configuration: Serialize + Clone + Sync + Send,
    C::State: Sync + Send + Clone,
{

    // let tracer = state.tracer.clone();

    // * Pass tracing middleware struct in via ServerState.
    // * Create OTel Trace Middleware w/ helper
    Router::new()
        .route("/capabilities", get(get_capabilities::<C>))
        .route("/health", get(get_health::<C>))
        .route("/metrics", get(get_metrics::<C>))
        .route("/schema", get(get_schema::<C>))
        .route("/query", post(post_query::<C>))
        .route("/explain", post(post_explain::<C>))
        .route("/mutation", post(post_mutation::<C>))
        .with_state(state)
        .layer(
            TraceLayer::new_for_http()
            // .on_response(|response: &Response<BoxBody>, latency: Duration, span: &Span| {
            //     // Using callback implementation as documented here: https://docs.rs/tower-http/0.4.1/tower_http/trace/index.html
            //     // Unfortunately, there's extensive work required to make the request available in the on_response callback.
            //     // Also, `in_span` is the tracer method documented, and this wants to wrap the execution, as opposed to append to the prior on_request trace.
            //     tracing::debug!("response generated in {:?}", latency)
            // })
        )
        // .layer(from_fn_with_state(tracer, otlp_middleware))
        // * Add as .layer
        .layer(response_with_trace_layer())
        // set up `TraceLayer` from tower-http
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

mod v2_compat;

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
use axum_extra::extract::WithRejection;
use clap::{Parser, Subcommand};
use prometheus::Registry;
use serde::Serialize;
use tower_http::{trace::TraceLayer, validate_request::ValidateRequestHeaderLayer};

use ndc_client::models::{
    CapabilitiesResponse, ErrorResponse, ExplainResponse, MutationRequest, MutationResponse,
    QueryRequest, QueryResponse, SchemaResponse,
};
use ndc_test::report;

use crate::check_health;
use crate::connector::Connector;
use crate::json_rejection::JsonRejection;
use crate::json_response::JsonResponse;
use crate::routes;
use crate::tracing::{init_tracing, make_span, on_response};

#[derive(Parser)]
struct CliArgs {
    #[command(subcommand)]
    command: Command,
}

#[derive(Clone, Subcommand)]
enum Command {
    #[command()]
    Serve(ServeCommand),
    #[command()]
    Test(TestCommand),
    #[command()]
    Replay(ReplayCommand),
    #[command()]
    CheckHealth(CheckHealthCommand),
}

#[derive(Clone, Parser)]
struct ServeCommand {
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_CONFIGURATION_DIRECTORY")]
    configuration: PathBuf,
    #[arg(long, value_name = "ENDPOINT", env = "OTEL_EXPORTER_OTLP_ENDPOINT")]
    otlp_endpoint: Option<String>,
    #[arg(
        long,
        value_name = "PORT",
        env = "HASURA_CONNECTOR_PORT",
        default_value_t = 8080
    )]
    port: Port,
    #[arg(long, value_name = "TOKEN", env = "HASURA_SERVICE_TOKEN_SECRET")]
    service_token_secret: Option<String>,
    #[arg(long, value_name = "NAME", env = "OTEL_SERVICE_NAME")]
    service_name: Option<String>,
    #[arg(long, env = "HASURA_ENABLE_V2_COMPATIBILITY")]
    enable_v2_compatibility: bool,
}

#[derive(Clone, Parser)]
struct TestCommand {
    #[arg(long, value_name = "SEED", env = "SEED")]
    seed: Option<String>,
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_CONFIGURATION_DIRECTORY")]
    configuration: PathBuf,
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_SNAPSHOTS_DIR")]
    snapshots_dir: Option<PathBuf>,
}

#[derive(Clone, Parser)]
struct ReplayCommand {
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_CONFIGURATION_DIRECTORY")]
    configuration: PathBuf,
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_SNAPSHOTS_DIR")]
    snapshots_dir: PathBuf,
}

#[derive(Clone, Parser)]
struct CheckHealthCommand {
    #[arg(long, value_name = "HOST")]
    host: Option<String>,
    #[arg(
        long,
        value_name = "PORT",
        env = "HASURA_CONNECTOR_PORT",
        default_value_t = 8080
    )]
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
pub async fn default_main<C: Connector + 'static>() -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::Configuration: Clone + Serialize,
    C::State: Clone,
{
    let CliArgs { command } = CliArgs::parse();

    match command {
        Command::Serve(serve_command) => serve::<C>(serve_command).await,
        Command::Test(test_command) => test::<C>(test_command).await,
        Command::Replay(replay_command) => replay::<C>(replay_command).await,
        Command::CheckHealth(check_health_command) => check_health(check_health_command).await,
    }
}

async fn serve<C: Connector + 'static>(
    serve_command: ServeCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    C::Configuration: Serialize + Clone,
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
pub async fn init_server_state<C: Connector>(
    config_directory: impl AsRef<Path> + Send,
) -> ServerState<C> {
    let configuration = C::parse_configuration(config_directory).await.unwrap();

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
    C::Configuration: Clone,
    C::State: Clone,
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
    C::Configuration: Clone + Serialize,
    C::State: Clone,
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

struct ConnectorAdapter<C: Connector> {
    configuration: C::Configuration,
    state: C::State,
}

#[async_trait]
impl<C: Connector> ndc_test::Connector for ConnectorAdapter<C> {
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

async fn test<C: Connector>(command: TestCommand) -> Result<(), Box<dyn Error + Send + Sync>> {
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

async fn replay<C: Connector>(command: ReplayCommand) -> Result<(), Box<dyn Error + Send + Sync>> {
    let connector = make_connector_adapter::<C>(command.configuration).await;
    let results = ndc_test::test_snapshots_in_directory(&connector, command.snapshots_dir).await;

    if !results.failures.is_empty() {
        println!();
        println!("{}", report(results));

        exit(1)
    }

    Ok(())
}

async fn make_connector_adapter<C: Connector>(configuration_path: PathBuf) -> ConnectorAdapter<C> {
    let configuration = C::parse_configuration(configuration_path).await.unwrap();

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

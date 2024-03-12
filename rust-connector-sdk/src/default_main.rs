use std::error::Error;
use std::net;
use std::path::{Path, PathBuf};

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
use tower_http::{trace::TraceLayer, validate_request::ValidateRequestHeaderLayer};

use ndc_client::models::{
    CapabilitiesResponse, ErrorResponse, ExplainResponse, MutationRequest, MutationResponse,
    QueryRequest, QueryResponse, SchemaResponse,
};

use crate::check_health;
use crate::connector::{Connector, ConnectorSetup};
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
    #[cfg(feature = "ndc-test")]
    Test(TestCommand),
    #[command()]
    #[cfg(feature = "ndc-test")]
    Replay(ReplayCommand),
    #[command()]
    #[cfg(feature = "ndc-test")]
    Bench(BenchCommand),
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
        value_name = "HOST IP",
        env = "HASURA_CONNECTOR_HOST",
        default_value_t = net::IpAddr::V4(net::Ipv4Addr::UNSPECIFIED),
    )]
    host: net::IpAddr,
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
struct BenchCommand {
    #[arg(long, value_name = "DIRECTORY", env = "HASURA_CONFIGURATION_DIRECTORY")]
    configuration: PathBuf,
    #[arg(
        long,
        value_name = "COUNT",
        help = "the number of samples to collect per test",
        default_value = "100"
    )]
    samples: u32,
    #[arg(
        long,
        value_name = "TOLERANCE",
        help = "tolerable deviation from previous report, in standard deviations from the mean"
    )]
    tolerance: Option<f64>,
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

impl<C: Connector> ServerState<C> {
    pub fn new(configuration: C::Configuration, state: C::State, metrics: Registry) -> Self {
        Self {
            configuration,
            state,
            metrics,
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
pub async fn default_main<Setup>() -> Result<(), Box<dyn Error + Send + Sync>>
where
    Setup: ConnectorSetup + Default,
    Setup::Connector: Connector + 'static,
    <Setup::Connector as Connector>::Configuration: Clone,
    <Setup::Connector as Connector>::State: Clone,
{
    default_main_with(Setup::default()).await
}

/// A default main function for a connector, with a non-default setup.
///
/// See [`default_main`] for further details.
pub async fn default_main_with<Setup>(setup: Setup) -> Result<(), Box<dyn Error + Send + Sync>>
where
    Setup: ConnectorSetup,
    Setup::Connector: Connector + 'static,
    <Setup::Connector as Connector>::Configuration: Clone,
    <Setup::Connector as Connector>::State: Clone,
{
    let CliArgs { command } = CliArgs::parse();

    match command {
        Command::Serve(serve_command) => serve(setup, serve_command).await,
        Command::CheckHealth(check_health_command) => check_health(check_health_command).await,
        #[cfg(feature = "ndc-test")]
        Command::Test(test_command) => ndc_test_commands::test(setup, test_command).await,
        #[cfg(feature = "ndc-test")]
        Command::Bench(bench_command) => ndc_test_commands::bench(setup, bench_command).await,
        #[cfg(feature = "ndc-test")]
        Command::Replay(replay_command) => ndc_test_commands::replay(setup, replay_command).await,
    }
}

async fn serve<Setup>(
    setup: Setup,
    serve_command: ServeCommand,
) -> Result<(), Box<dyn Error + Send + Sync>>
where
    Setup: ConnectorSetup,
    Setup::Connector: Connector + 'static,
    <Setup::Connector as Connector>::Configuration: Clone,
    <Setup::Connector as Connector>::State: Clone,
{
    init_tracing(&serve_command.service_name, &serve_command.otlp_endpoint)
        .expect("Unable to initialize tracing");

    let server_state = init_server_state(setup, serve_command.configuration).await?;

    let router = create_router::<Setup::Connector>(
        server_state.clone(),
        serve_command.service_token_secret.clone(),
    );

    let address = net::SocketAddr::new(serve_command.host, serve_command.port);
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
pub async fn init_server_state<Setup: ConnectorSetup>(
    setup: Setup,
    config_directory: impl AsRef<Path> + Send,
) -> Result<ServerState<Setup::Connector>, Box<dyn Error + Send + Sync>> {
    let mut metrics = Registry::new();
    let configuration = setup.parse_configuration(config_directory).await?;
    let state = setup.try_init_state(&configuration, &mut metrics).await?;
    Ok(ServerState::new(configuration, state, metrics))
}

pub fn create_router<C>(state: ServerState<C>, service_token_secret: Option<String>) -> Router
where
    C: Connector + 'static,
    C::Configuration: Clone,
    C::State: Clone,
{
    let router = Router::new()
        .route("/capabilities", get(get_capabilities::<C>))
        .route("/health", get(get_health::<C>))
        .route("/metrics", get(get_metrics::<C>))
        .route("/schema", get(get_schema::<C>))
        .route("/query", post(post_query::<C>))
        .route("/query/explain", post(post_query_explain::<C>))
        .route("/mutation", post(post_mutation::<C>))
        .route("/mutation/explain", post(post_mutation_explain::<C>))
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

async fn post_query_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    WithRejection(Json(request), _): WithRejection<Json<QueryRequest>, JsonRejection>,
) -> Result<JsonResponse<ExplainResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_query_explain::<C>(&state.configuration, &state.state, request).await
}

async fn post_mutation_explain<C: Connector>(
    State(state): State<ServerState<C>>,
    WithRejection(Json(request), _): WithRejection<Json<MutationRequest>, JsonRejection>,
) -> Result<JsonResponse<ExplainResponse>, (StatusCode, Json<ErrorResponse>)> {
    routes::post_mutation_explain::<C>(&state.configuration, &state.state, request).await
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

#[cfg(feature = "ndc-test")]
mod ndc_test_commands {
    use super::{BenchCommand, Connector, ConnectorSetup};
    use crate::json_response::JsonResponse;
    use async_trait::async_trait;
    use ndc_test::reporter::{ConsoleReporter, TestResults};
    use prometheus::Registry;
    use std::error::Error;
    use std::path::PathBuf;
    use std::process::exit;

    struct ConnectorAdapter<C: Connector> {
        configuration: C::Configuration,
        state: C::State,
    }

    #[async_trait(?Send)]
    impl<C: Connector> ndc_test::connector::Connector for ConnectorAdapter<C> {
        async fn get_capabilities(
            &self,
        ) -> Result<ndc_client::models::CapabilitiesResponse, ndc_test::error::Error> {
            C::get_capabilities()
                .await
                .into_value::<Box<dyn std::error::Error + Send + Sync>>()
                .map_err(|err| ndc_test::error::Error::OtherError(err))
        }

        async fn get_schema(
            &self,
        ) -> Result<ndc_client::models::SchemaResponse, ndc_test::error::Error> {
            match C::get_schema(&self.configuration).await {
                Ok(response) => response
                    .into_value::<Box<dyn std::error::Error + Send + Sync>>()
                    .map_err(|err| ndc_test::error::Error::OtherError(err)),
                Err(err) => Err(ndc_test::error::Error::OtherError(err.into())),
            }
        }

        async fn query(
            &self,
            request: ndc_client::models::QueryRequest,
        ) -> Result<ndc_client::models::QueryResponse, ndc_test::error::Error> {
            match C::query(&self.configuration, &self.state, request)
                .await
                .and_then(JsonResponse::into_value)
            {
                Ok(response) => Ok(response),
                Err(err) => Err(ndc_test::error::Error::OtherError(err.into())),
            }
        }

        async fn mutation(
            &self,
            request: ndc_client::models::MutationRequest,
        ) -> Result<ndc_client::models::MutationResponse, ndc_test::error::Error> {
            match C::mutation(&self.configuration, &self.state, request)
                .await
                .and_then(JsonResponse::into_value)
            {
                Ok(response) => Ok(response),
                Err(err) => Err(ndc_test::error::Error::OtherError(err.into())),
            }
        }
    }

    pub(super) async fn test<Setup: super::ConnectorSetup>(
        setup: Setup,
        command: super::TestCommand,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let test_configuration = ndc_test::configuration::TestConfiguration {
            seed: command.seed.map(|s| s.as_bytes().try_into()).transpose()?,
            snapshots_dir: command.snapshots_dir,
            gen_config: ndc_test::configuration::TestGenerationConfiguration::default(),
        };

        let connector = make_connector_adapter(setup, command.configuration).await?;
        let mut reporter = (ConsoleReporter::new(), TestResults::default());

        ndc_test::test_connector(&test_configuration, &connector, &mut reporter).await;

        if !reporter.1.failures.is_empty() {
            println!();
            println!("{}", reporter.1.report());

            exit(1)
        }

        Ok(())
    }

    pub(super) async fn replay<Setup: super::ConnectorSetup>(
        setup: Setup,
        command: super::ReplayCommand,
    ) -> Result<(), Box<dyn Error + Send + Sync>> {
        let connector = make_connector_adapter(setup, command.configuration).await?;
        let mut reporter = (ConsoleReporter::new(), TestResults::default());

        ndc_test::test_snapshots_in_directory(&connector, &mut reporter, command.snapshots_dir)
            .await;

        if !reporter.1.failures.is_empty() {
            println!();
            println!("{}", reporter.1.report());

            exit(1)
        }

        Ok(())
    }

    pub(super) async fn bench<Setup: ConnectorSetup>(
        setup: Setup,
        command: BenchCommand,
    ) -> Result<(), Box<dyn Error + Send + Sync>> {
        let configuration = ndc_test::ReportConfiguration {
            samples: command.samples,
            tolerance: command.tolerance,
        };

        let connector = make_connector_adapter(setup, command.configuration).await?;
        let mut reporter = (ConsoleReporter::new(), TestResults::default());

        let reports = ndc_test::bench_snapshots_in_directory(
            &configuration,
            &connector,
            &mut reporter,
            command.snapshots_dir,
        )
        .await
        .map_err(|e| e.to_string())?;

        println!();
        println!("{}", ndc_test::benchmark_report(&configuration, reports));

        if !reporter.1.failures.is_empty() {
            exit(1);
        }

        Ok(())
    }

    async fn make_connector_adapter<Setup: ConnectorSetup>(
        setup: Setup,
        configuration_path: PathBuf,
    ) -> Result<ConnectorAdapter<Setup::Connector>, Box<dyn Error + Send + Sync>> {
        let mut metrics = Registry::new();
        let configuration = setup.parse_configuration(configuration_path).await?;
        let state = setup.try_init_state(&configuration, &mut metrics).await?;
        Ok(ConnectorAdapter {
            configuration,
            state,
        })
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

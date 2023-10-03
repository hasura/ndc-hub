use opentelemetry::{global, sdk::propagation::TraceContextPropagator};
use opentelemetry_api::KeyValue;
use opentelemetry_otlp::{WithExportConfig, OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT};
use opentelemetry_sdk::trace::Sampler;
use std::env;
use std::error::Error;
use tracing_subscriber::EnvFilter;

use axum::body::{Body, BoxBody};
use http::{Request, Response, Uri};
use opentelemetry_http::HeaderExtractor;
use std::time::Duration;
use tracing::{Level, Span};
use tracing_opentelemetry::OpenTelemetrySpanExt;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;

pub fn init_tracing(
    service_name: &Option<String>,
    otlp_endpoint: &Option<String>,
) -> Result<(), Box<dyn Error>> {
    global::set_text_map_propagator(TraceContextPropagator::new());

    let service_name = service_name
        .clone()
        .unwrap_or(env!("CARGO_PKG_NAME").to_string());

    let tracer = opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(
            opentelemetry_otlp::new_exporter().tonic().with_endpoint(
                otlp_endpoint
                    .clone()
                    .unwrap_or(OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT.into()),
            ),
        )
        .with_trace_config(
            opentelemetry::sdk::trace::config()
                .with_resource(opentelemetry::sdk::Resource::new(vec![
                    KeyValue::new(
                        opentelemetry_semantic_conventions::resource::SERVICE_NAME,
                        service_name,
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
// Custom function for creating request-level spans
// tracing crate requires all fields to be defined at creation time, so any fields that will be set
// later should be defined as Empty
pub fn make_span(request: &Request<Body>) -> Span {
    let span = tracing::span!(
        Level::INFO,
        "request",
        method = %request.method(),
        uri = %request.uri(),
        version = ?request.version(),
        deployment_id = extract_deployment_id(request.uri()),
        status = tracing::field::Empty,
        latency = tracing::field::Empty,
    );

    // Get parent trace id from headers, if available
    // This uses OTel extension set_parent rather than setting field directly on the span to ensure
    // it works no matter which propagator is configured
    let parent_context = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(request.headers()))
    });
    span.set_parent(parent_context);

    return span;
}

// Rough implementation of extracting deployment ID from URI. Regex might be better?
fn extract_deployment_id(uri: &Uri) -> &str {
    let path = uri.path();
    let mut parts = path.split('/').filter(|x| !x.is_empty());
    let _ = parts.next().unwrap_or_default();
    parts.next().unwrap_or_else(|| "unknown")
}

// Custom function for adding information to request-level span that is only available at response time.
pub fn on_response(response: &Response<BoxBody>, latency: Duration, span: &Span) {
    span.record("status", tracing::field::display(response.status()));
    span.record("latency", tracing::field::display(latency.as_nanos()));
}

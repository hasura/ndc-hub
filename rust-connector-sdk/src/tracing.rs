use std::env;
use std::error::Error;
use std::time::Duration;

use axum::body::{Body, BoxBody};
use http::{Request, Response};
use opentelemetry::{global, sdk::propagation::TraceContextPropagator};
use opentelemetry_api::KeyValue;
use opentelemetry_http::HeaderExtractor;
use opentelemetry_otlp::{WithExportConfig, OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT};
use opentelemetry_sdk::trace::Sampler;
use tracing::Span;
use tracing_opentelemetry::OpenTelemetrySpanExt;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

pub fn init_tracing(
    service_name: &Option<String>,
    otlp_endpoint: &Option<String>,
) -> Result<(), Box<dyn Error + Send + Sync>> {
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
    use opentelemetry::trace::TraceContextExt;

    let span = tracing::info_span!(
        "request",
        method = %request.method(),
        uri = %request.uri(),
        version = ?request.version(),
        status = tracing::field::Empty,
        latency = tracing::field::Empty,
    );

    // Get parent trace id from headers, if available
    // This uses OTel extension set_parent rather than setting field directly on the span to ensure
    // it works no matter which propagator is configured
    let parent_context = global::get_text_map_propagator(|propagator| {
        propagator.extract(&HeaderExtractor(request.headers()))
    });
    // if there is no parent span ID, we get something nonsensical, so we need to validate it
    // (yes, this is hilarious)
    let parent_context_span = parent_context.span();
    let parent_context_span_context = parent_context_span.span_context();
    if parent_context_span_context.is_valid() {
        span.set_parent(parent_context);
    }

    span
}

// Custom function for adding information to request-level span that is only available at response time.
pub fn on_response(response: &Response<BoxBody>, latency: Duration, span: &Span) {
    span.record("status", tracing::field::display(response.status()));
    span.record("latency", tracing::field::display(latency.as_nanos()));
}

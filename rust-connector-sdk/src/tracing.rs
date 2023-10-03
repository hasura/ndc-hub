use opentelemetry::{global, sdk::propagation::TraceContextPropagator};
use opentelemetry_api::KeyValue;
use opentelemetry_otlp::{WithExportConfig, OTEL_EXPORTER_OTLP_ENDPOINT_DEFAULT};
use opentelemetry_sdk::trace::Sampler;
use std::env;
use std::error::Error;
use tracing_subscriber::{prelude::*, EnvFilter};

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

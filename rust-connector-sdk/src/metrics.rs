//! Some metrics setup and update.

use crate::json_response::JsonResponse;
use axum::{http::StatusCode, Json};
use ndc_client::models::ErrorResponse;
use prometheus;

/// The collection of some metrics exposed through the `/metrics` endpoint.
#[derive(Debug, Clone)]
pub struct Metrics {
    status_codes: StatusCodeMetrics,
}

/// The collection of some metrics exposed through the `/metrics` endpoint.
#[derive(Debug, Clone)]
pub struct StatusCodeMetrics {
    total_200: prometheus::IntCounter,
    total_400: prometheus::IntCounter,
    total_403: prometheus::IntCounter,
    total_409: prometheus::IntCounter,
    total_500: prometheus::IntCounter,
    total_501: prometheus::IntCounter,
}

impl Metrics {
    /// Set up counters and gauges used to produce Prometheus metrics
    pub fn initialize(
        metrics_registry: &mut prometheus::Registry,
    ) -> Result<Self, prometheus::Error> {
        let total_200 = add_int_counter_metric(
            metrics_registry,
            "status_code_200",
            "Total number of 200 status codes returned.",
        )?;

        let total_400 = add_int_counter_metric(
            metrics_registry,
            "status_code_400",
            "Total number of 400 status codes returned.",
        )?;

        let total_403 = add_int_counter_metric(
            metrics_registry,
            "status_code_403",
            "Total number of 403 status codes returned.",
        )?;

        let total_409 = add_int_counter_metric(
            metrics_registry,
            "status_code_409",
            "Total number of 409 status codes returned.",
        )?;

        let total_500 = add_int_counter_metric(
            metrics_registry,
            "status_code_500",
            "Total number of 500 status codes returned.",
        )?;

        let total_501 = add_int_counter_metric(
            metrics_registry,
            "status_code_501",
            "Total number of 501 status codes returned.",
        )?;

        let status_codes = StatusCodeMetrics {
            total_200,
            total_400,
            total_403,
            total_409,
            total_500,
            total_501,
        };

        Ok(Self { status_codes })
    }

    /// record a status code from an api result.
    pub fn record_status<T>(
        &self,
        result: Result<JsonResponse<T>, (StatusCode, Json<ErrorResponse>)>,
    ) -> Result<JsonResponse<T>, (StatusCode, Json<ErrorResponse>)> {
        match result {
            Ok(result) => {
                self.record_status_code(StatusCode::OK);
                Ok(result)
            }
            Err((status_code, result)) => {
                self.record_status_code(status_code);
                Err((status_code, result))
            }
        }
    }

    fn record_status_code(&self, status_code: StatusCode) {
        match status_code {
            StatusCode::OK => self.record_200(),
            StatusCode::BAD_REQUEST => self.record_400(),
            StatusCode::FORBIDDEN => self.record_403(),
            StatusCode::CONFLICT => self.record_409(),
            StatusCode::INTERNAL_SERVER_ERROR => self.record_500(),
            StatusCode::NOT_IMPLEMENTED => self.record_501(),
            _ => (),
        }
    }

    fn record_200(&self) {
        self.status_codes.total_200.inc()
    }
    fn record_400(&self) {
        self.status_codes.total_400.inc()
    }
    fn record_403(&self) {
        self.status_codes.total_403.inc()
    }
    fn record_409(&self) {
        self.status_codes.total_409.inc()
    }
    fn record_500(&self) {
        self.status_codes.total_500.inc()
    }
    fn record_501(&self) {
        self.status_codes.total_501.inc()
    }
}

/// Create a new int counter metric and register it with the provided Prometheus Registry
fn add_int_counter_metric(
    metrics_registry: &mut prometheus::Registry,
    metric_name: &str,
    metric_description: &str,
) -> Result<prometheus::IntCounter, prometheus::Error> {
    let int_counter =
        prometheus::IntCounter::with_opts(prometheus::Opts::new(metric_name, metric_description))?;
    register_collector(metrics_registry, int_counter)
}

/// Register a new collector with the registry, and returns it for later use.
fn register_collector<Collector: prometheus::core::Collector + std::clone::Clone + 'static>(
    metrics_registry: &mut prometheus::Registry,
    collector: Collector,
) -> Result<Collector, prometheus::Error> {
    metrics_registry.register(Box::new(collector.clone()))?;
    Ok(collector)
}

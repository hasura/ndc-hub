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
    total_1xx: prometheus::IntCounter,
    total_2xx: prometheus::IntCounter,
    total_3xx: prometheus::IntCounter,
    total_4xx: prometheus::IntCounter,
    total_5xx: prometheus::IntCounter,
}

impl Metrics {
    /// Set up counters and gauges used to produce Prometheus metrics
    pub fn initialize(
        connector_name: String,
        metrics_registry: &mut prometheus::Registry,
    ) -> Result<Self, prometheus::Error> {
        // Transform the connector name so it is a valid prometheus metric name
        // <https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels>
        let connector_name: String = connector_name
            .chars()
            .filter_map(|c| {
                if c == '-' {
                    Some('_')
                } else if c.is_ascii_alphanumeric() {
                    Some(c)
                } else {
                    None
                }
            })
            .collect();

        let total_200 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_200_total_count", connector_name).as_str(),
            "Total number of 200 status codes returned.",
        )?;

        let total_400 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_400_total_count", connector_name).as_str(),
            "Total number of 400 status codes returned.",
        )?;

        let total_403 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_403_total_count", connector_name).as_str(),
            "Total number of 403 status codes returned.",
        )?;

        let total_409 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_409_total_count", connector_name).as_str(),
            "Total number of 409 status codes returned.",
        )?;

        let total_500 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_500_total_count", connector_name).as_str(),
            "Total number of 500 status codes returned.",
        )?;

        let total_501 = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_501_total_count", connector_name).as_str(),
            "Total number of 501 status codes returned.",
        )?;

        let total_1xx = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_1xx_total_count", connector_name).as_str(),
            "Total number of 1xx status codes returned.",
        )?;

        let total_2xx = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_2xx_total_count", connector_name).as_str(),
            "Total number of 2xx status codes returned.",
        )?;

        let total_3xx = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_3xx_total_count", connector_name).as_str(),
            "Total number of 3xx status codes returned.",
        )?;

        let total_4xx = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_4xx_total_count", connector_name).as_str(),
            "Total number of 4xx status codes returned.",
        )?;

        let total_5xx = add_int_counter_metric(
            metrics_registry,
            format!("hasura_{}_status_code_5xx_total_count", connector_name).as_str(),
            "Total number of 5xx status codes returned.",
        )?;

        let status_codes = StatusCodeMetrics {
            total_200,
            total_400,
            total_403,
            total_409,
            total_500,
            total_501,
            total_1xx,
            total_2xx,
            total_3xx,
            total_4xx,
            total_5xx,
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

    /// record a status code metric into individual and/or group buckets.
    fn record_status_code(&self, status_code: StatusCode) {
        // We record the status codes used by ndc-spec:
        // <https://hasura.github.io/ndc-spec/specification/error-handling.html>
        // and group all others into buckets.
        match status_code {
            StatusCode::OK => self.record_200(),
            StatusCode::BAD_REQUEST => self.record_400(),
            StatusCode::FORBIDDEN => self.record_403(),
            StatusCode::CONFLICT => self.record_409(),
            StatusCode::INTERNAL_SERVER_ERROR => self.record_500(),
            StatusCode::NOT_IMPLEMENTED => self.record_501(),
            status => {
                let code = status.as_u16();
                if code < 200 {
                    self.record_1xx();
                } else if code < 300 {
                    self.record_2xx();
                } else if code < 400 {
                    self.record_3xx();
                } else if code < 500 {
                    self.record_4xx();
                } else {
                    self.record_5xx();
                }
            }
        }
    }

    fn record_200(&self) {
        self.status_codes.total_200.inc();
        self.record_2xx();
    }
    fn record_400(&self) {
        self.status_codes.total_400.inc();
        self.record_4xx();
    }
    fn record_403(&self) {
        self.status_codes.total_403.inc();
        self.record_4xx();
    }
    fn record_409(&self) {
        self.status_codes.total_409.inc();
        self.record_4xx();
    }
    fn record_500(&self) {
        self.status_codes.total_500.inc();
        self.record_5xx();
    }
    fn record_501(&self) {
        self.status_codes.total_501.inc();
        self.record_5xx();
    }
    fn record_1xx(&self) {
        self.status_codes.total_1xx.inc();
    }
    fn record_2xx(&self) {
        self.status_codes.total_2xx.inc();
    }
    fn record_3xx(&self) {
        self.status_codes.total_3xx.inc();
    }
    fn record_4xx(&self) {
        self.status_codes.total_4xx.inc();
    }
    fn record_5xx(&self) {
        self.status_codes.total_5xx.inc();
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

use async_trait::async_trait;
use ndc_client::models;
use serde::Serialize;
use std::error::Error;
use thiserror::Error;

use crate::json_response::JsonResponse;

pub mod example;

/// Errors which occur when trying to validate connector
/// configuration.
///
/// See [`Connector::validate_raw_configuration`].
#[derive(Debug, Error)]
pub enum ValidateError {
    #[error("error validating configuration: {0:?}")]
    ValidateError(Vec<InvalidRange>),
}

#[derive(Debug, Clone, Serialize)]
pub struct InvalidRange {
    pub path: Vec<KeyOrIndex>,
    pub message: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(untagged)]
pub enum KeyOrIndex {
    Key(String),
    Index(u32),
}

/// Errors which occur when trying to validate connector
/// configuration.
///
/// See [`Connector::update_configuration`].
#[derive(Debug, Error)]
pub enum UpdateConfigurationError {
    #[error("error validating configuration: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when trying to initialize connector
/// state.
///
/// See [`Connector::try_init_state`].
#[derive(Debug, Error)]
pub enum InitializationError {
    #[error("error initializing connector state: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when trying to update metrics.
///
/// See [`Connector::fetch_metrics`].
#[derive(Debug, Error)]
pub enum FetchMetricsError {
    #[error("error fetching metrics: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when checking connector health.
///
/// See [`Connector::health_check`].
#[derive(Debug, Error)]
pub enum HealthError {
    #[error("error checking health status: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when retrieving the connector schema.
///
/// See [`Connector::get_schema`].
#[derive(Debug, Error)]
pub enum SchemaError {
    #[error("error retrieving the schema: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when executing a query.
///
/// See [`Connector::query`].
#[derive(Debug, Error)]
pub enum QueryError {
    /// The request was invalid or did not match the
    /// requirements of the specification. This indicates
    /// an error with the client.
    #[error("invalid request: {0}")]
    InvalidRequest(String),
    /// The request relies on an unsupported feature or
    /// capability. This may indicate an error with the client,
    /// or just an unimplemented feature.
    #[error("unsupported operation: {0}")]
    UnsupportedOperation(String),
    #[error("error executing query: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when explaining a query.
///
/// See [`Connector::explain`].
#[derive(Debug, Error)]
pub enum ExplainError {
    /// The request was invalid or did not match the
    /// requirements of the specification. This indicates
    /// an error with the client.
    #[error("invalid request: {0}")]
    InvalidRequest(String),
    /// The request relies on an unsupported feature or
    /// capability. This may indicate an error with the client,
    /// or just an unimplemented feature.
    #[error("unsupported operation: {0}")]
    UnsupportedOperation(String),
    #[error("error explaining query: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Errors which occur when executing a mutation.
///
/// See [`Connector::mutation`].
#[derive(Debug, Error)]
pub enum MutationError {
    /// The request was invalid or did not match the
    /// requirements of the specification. This indicates
    /// an error with the client.
    #[error("invalid request: {0}")]
    InvalidRequest(String),
    /// The request relies on an unsupported feature or
    /// capability. This may indicate an error with the client,
    /// or just an unimplemented feature.
    #[error("unsupported operation: {0}")]
    UnsupportedOperation(String),
    /// The request would result in a conflicting state
    /// in the underlying data store.
    #[error("mutation would result in a conflicting state: {0}")]
    Conflict(String),
    /// The request would violate a constraint in the
    /// underlying data store.
    #[error("mutation violates constraint: {0}")]
    ConstraintNotMet(String),
    #[error("error executing mutation: {0}")]
    Other(#[from] Box<dyn Error + Send + Sync>),
}

/// Connectors using this library should implement this trait.
///
///
/// It provides methods which implement the standard endpoints
/// defined by the specification: capabilities, schema, query, mutation
/// and explain.
///
/// In addition, it introduces names for types to manage
/// state and configuration (if any), and provides any necessary context
/// for observability purposes (metrics, logging and tracing).
///
/// ## Configuration
///
/// Connectors encapsulate data sources, and likely require configuration
/// (connection strings, web service tokens, etc.). The NDC specification
/// does not discuss this sort of configuration, because it is an
/// implementation detail of a specific connector, but it is useful to
/// adopt a convention here for simplified configuration management.
///
/// Configuration is specified as JSON, validated, and stored in a binary
/// format.
///
/// This trait defines two types for managing configuration:
///
/// - [`Connector::RawConfiguration`] defines the type of unvalidated, raw
///   configuration.
/// - [`Connector::Configuration`] defines the type of validated
///   configuration. Ideally, invalid configuration should not be representable
///   in this form.
///
/// ## State
///
/// In addition to configuration, this trait defines a type for state management:
///
/// - [`Connector::State`] defines the type of any unserializable runtime state.
///
/// State is distinguished from configuration in that it is not provided directly by
/// the user, and would not ordinarily be serializable. For example, a connection string
/// would be configuration, but a connection pool object created from that
/// connection string would be state.
#[async_trait]
pub trait Connector {
    /// The type of unvalidated, raw configuration, as provided by the user.
    type RawConfiguration;
    /// The type of validated configuration
    type Configuration;
    /// The type of unserializable state
    type State;

    fn connector_name() -> String;

    fn make_empty_configuration() -> Self::RawConfiguration;

    async fn update_configuration(
        config: Self::RawConfiguration,
    ) -> Result<Self::RawConfiguration, UpdateConfigurationError>;

    /// Validate the raw configuration provided by the user,
    /// returning a configuration error or a validated [`Connector::Configuration`].
    async fn validate_raw_configuration(
        configuration: Self::RawConfiguration,
    ) -> Result<Self::Configuration, ValidateError>;

    /// Initialize the connector's in-memory state.
    ///
    /// For example, any connection pools, prepared queries,
    /// or other managed resources would be allocated here.
    ///
    /// In addition, this function should register any
    /// connector-specific metrics with the metrics registry.
    async fn try_init_state(
        configuration: &Self::Configuration,
        metrics: &mut prometheus::Registry,
    ) -> Result<Self::State, InitializationError>;

    /// Update any metrics from the state
    ///
    /// Note: some metrics can be updated directly, and do not
    /// need to be updated here. This function can be useful to
    /// query metrics which cannot be updated directly, e.g.
    /// the number of idle connections in a connection pool
    /// can be polled but not updated directly.
    fn fetch_metrics(
        configuration: &Self::Configuration,
        state: &Self::State,
    ) -> Result<(), FetchMetricsError>;

    /// Check the health of the connector.
    ///
    /// For example, this function should check that the connector
    /// is able to reach its data source over the network.
    async fn health_check(
        configuration: &Self::Configuration,
        state: &Self::State,
    ) -> Result<(), HealthError>;

    /// Get the connector's capabilities.
    ///
    /// This function implements the [capabilities endpoint](https://hasura.github.io/ndc-spec/specification/capabilities.html)
    /// from the NDC specification.
    async fn get_capabilities() -> JsonResponse<models::CapabilitiesResponse>;

    /// Get the connector's schema.
    ///
    /// This function implements the [schema endpoint](https://hasura.github.io/ndc-spec/specification/schema/index.html)
    /// from the NDC specification.
    async fn get_schema(
        configuration: &Self::Configuration,
    ) -> Result<JsonResponse<models::SchemaResponse>, SchemaError>;

    /// Explain a query by creating an execution plan
    ///
    /// This function implements the [explain endpoint](https://hasura.github.io/ndc-spec/specification/explain.html)
    /// from the NDC specification.
    async fn explain(
        configuration: &Self::Configuration,
        state: &Self::State,
        request: models::QueryRequest,
    ) -> Result<JsonResponse<models::ExplainResponse>, ExplainError>;

    /// Execute a mutation
    ///
    /// This function implements the [mutation endpoint](https://hasura.github.io/ndc-spec/specification/mutations/index.html)
    /// from the NDC specification.
    async fn mutation(
        configuration: &Self::Configuration,
        state: &Self::State,
        request: models::MutationRequest,
    ) -> Result<JsonResponse<models::MutationResponse>, MutationError>;

    /// Execute a query
    ///
    /// This function implements the [query endpoint](https://hasura.github.io/ndc-spec/specification/queries/index.html)
    /// from the NDC specification.
    async fn query(
        configuration: &Self::Configuration,
        state: &Self::State,
        request: models::QueryRequest,
    ) -> Result<JsonResponse<models::QueryResponse>, QueryError>;
}

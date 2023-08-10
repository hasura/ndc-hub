use std::collections::BTreeMap;

use async_trait::async_trait;
use tracing::info_span;
use tracing::Instrument;

use super::*;

#[derive(Clone, Default)]
pub struct Example {}

#[async_trait]
impl Connector for Example {
    type RawConfiguration = ();
    type Configuration = ();
    type State = ();

    fn make_empty_configuration() -> Self::RawConfiguration {
        ()
    }

    async fn update_configuration(
        _config: &Self::RawConfiguration,
    ) -> Result<Self::RawConfiguration, UpdateConfigurationError> {
        Ok(())
    }

    async fn validate_raw_configuration(
        _configuration: &Self::Configuration,
    ) -> Result<Self::Configuration, ValidateError> {
        Ok(())
    }

    async fn try_init_state(
        _configuration: &Self::Configuration,
        _metrics: &mut prometheus::Registry,
    ) -> Result<Self::State, InitializationError> {
        Ok(())
    }

    fn fetch_metrics(
        _configuration: &Self::Configuration,
        _state: &Self::State,
    ) -> Result<(), FetchMetricsError> {
        Ok(())
    }

    async fn health_check(
        _configuration: &Self::Configuration,
        _state: &Self::State,
    ) -> Result<(), HealthError> {
        Ok(())
    }

    async fn get_capabilities() -> models::CapabilitiesResponse {
        models::CapabilitiesResponse {
            versions: "^1.0.0".into(),
            capabilities: models::Capabilities {
                explain: None,
                relationships: None,
                mutations: None,
                query: Some(models::QueryCapabilities {
                    foreach: None,
                    order_by_aggregate: None,
                    relation_comparisons: None,
                }),
            },
        }
    }

    async fn get_schema(
        _configuration: &Self::Configuration,
    ) -> Result<models::SchemaResponse, SchemaError> {
        async {
            info_span!("inside tracing example");
            return ();
        }
        .instrument(info_span!("tracing example"))
        .await;

        Ok(models::SchemaResponse {
            collections: vec![],
            functions: vec![],
            procedures: vec![],
            object_types: BTreeMap::new(),
            scalar_types: BTreeMap::new(),
        })
    }

    async fn explain(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::QueryRequest,
    ) -> Result<models::ExplainResponse, ExplainError> {
        todo!()
    }

    async fn mutation(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::MutationRequest,
    ) -> Result<models::MutationResponse, MutationError> {
        todo!()
    }

    async fn query(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::QueryRequest,
    ) -> Result<models::QueryResponse, QueryError> {
        todo!()
    }
}

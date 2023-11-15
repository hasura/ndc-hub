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

    fn make_empty_configuration() -> Self::RawConfiguration {}

    async fn update_configuration(
        _config: Self::RawConfiguration,
    ) -> Result<Self::RawConfiguration, UpdateConfigurationError> {
        Ok(())
    }

    async fn validate_raw_configuration(
        _configuration: Self::Configuration,
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

    async fn get_capabilities() -> JsonResponse<models::CapabilitiesResponse> {
        models::CapabilitiesResponse {
            versions: "^0.1.0".into(),
            capabilities: models::Capabilities {
                explain: None,
                relationships: None,
                query: models::QueryCapabilities {
                    variables: None,
                    aggregates: None,
                },
            },
        }
        .into()
    }

    async fn get_schema(
        _configuration: &Self::Configuration,
    ) -> Result<JsonResponse<models::SchemaResponse>, SchemaError> {
        async {
            info_span!("inside tracing example");
        }
        .instrument(info_span!("tracing example"))
        .await;

        Ok(models::SchemaResponse {
            collections: vec![],
            functions: vec![],
            procedures: vec![],
            object_types: BTreeMap::new(),
            scalar_types: BTreeMap::new(),
        }
        .into())
    }

    async fn explain(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::QueryRequest,
    ) -> Result<JsonResponse<models::ExplainResponse>, ExplainError> {
        todo!()
    }

    async fn mutation(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::MutationRequest,
    ) -> Result<JsonResponse<models::MutationResponse>, MutationError> {
        todo!()
    }

    async fn query(
        _configuration: &Self::Configuration,
        _state: &Self::State,
        _request: models::QueryRequest,
    ) -> Result<JsonResponse<models::QueryResponse>, QueryError> {
        todo!()
    }
}

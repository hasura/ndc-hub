use axum::{http::StatusCode, Json};
use ndc_client::models::{self, CapabilitiesResponse};
use prometheus::{Encoder, Registry, TextEncoder};

use crate::connector::{Connector, HealthError};

pub fn get_metrics<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    metrics: Registry,
) -> Result<String, StatusCode> {
    let mut buffer = vec![];
    let encoder = TextEncoder::new();

    // Ask the connector to update all its metrics
    C::fetch_metrics(configuration, state).map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let metric_families = metrics.gather();
    encoder
        .encode(&metric_families, &mut buffer)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    String::from_utf8(buffer).map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

pub async fn get_capabilities<C: Connector>() -> Json<CapabilitiesResponse> {
    Json(C::get_capabilities().await)
}

pub async fn get_health<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
) -> StatusCode {
    // the context extractor will error if the deployment can't be found.
    match C::health_check(&configuration, &state).await {
        Ok(_) => StatusCode::NO_CONTENT,
        Err(HealthError::Other(_)) => {
            // TODO: log error
            StatusCode::SERVICE_UNAVAILABLE
        }
    }
}

pub async fn get_schema<C: Connector>(
    configuration: &C::Configuration,
) -> Result<Json<models::SchemaResponse>, StatusCode> {
    Ok(Json(C::get_schema(&configuration).await.map_err(
        |e| match e {
            crate::connector::SchemaError::Other(_) => {
                // TODO: log error
                StatusCode::INTERNAL_SERVER_ERROR
            }
        },
    )?))
}

pub async fn post_explain<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::QueryRequest>,
) -> Result<Json<models::ExplainResponse>, StatusCode> {
    Ok(Json(
        C::explain(&configuration, &state, request)
            .await
            .map_err(|e| match e {
                crate::connector::ExplainError::Other(_) => {
                    // TODO: log error
                    StatusCode::INTERNAL_SERVER_ERROR
                }
                crate::connector::ExplainError::InvalidRequest(_) => {
                    // TODO: log error
                    StatusCode::BAD_REQUEST
                }
                crate::connector::ExplainError::UnsupportedOperation(_) => {
                    // TODO: log error
                    StatusCode::NOT_IMPLEMENTED
                }
            })?,
    ))
}

pub async fn post_mutation<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::MutationRequest>,
) -> Result<Json<models::MutationResponse>, StatusCode> {
    Ok(Json(
        C::mutation(&configuration, &state, request)
            .await
            .map_err(|e| match e {
                crate::connector::MutationError::Other(_) => {
                    // TODO: log error
                    StatusCode::INTERNAL_SERVER_ERROR
                }
                crate::connector::MutationError::InvalidRequest(_) => {
                    // TODO: log error
                    StatusCode::BAD_REQUEST
                }
                crate::connector::MutationError::UnsupportedOperation(_) => {
                    // TODO: log error
                    StatusCode::NOT_IMPLEMENTED
                }
                crate::connector::MutationError::Conflict(_) => {
                    // TODO: log error
                    StatusCode::CONFLICT
                }
                crate::connector::MutationError::ConstraintNotMet(_) => {
                    // TODO: log error
                    StatusCode::FORBIDDEN
                }
            })?,
    ))
}

pub async fn post_query<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::QueryRequest>,
) -> Result<Json<models::QueryResponse>, StatusCode> {
    Ok(Json(
        C::query(&configuration, &state, request)
            .await
            .map_err(|e| match e {
                crate::connector::QueryError::Other(_) => {
                    // TODO: log error
                    StatusCode::INTERNAL_SERVER_ERROR
                }
                crate::connector::QueryError::InvalidRequest(_) => {
                    // TODO: log error
                    StatusCode::BAD_REQUEST
                }
                crate::connector::QueryError::UnsupportedOperation(_) => {
                    // TODO: log error
                    StatusCode::NOT_IMPLEMENTED
                }
            })?,
    ))
}

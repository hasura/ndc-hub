use axum::{http::StatusCode, Json};
use ndc_client::models;
use prometheus::{Registry, TextEncoder};

use crate::connector::{Connector, HealthError};

pub fn get_metrics<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    metrics: Registry,
) -> Result<String, (StatusCode, Json<models::ErrorResponse>)> {
    let encoder = TextEncoder::new();

    // Ask the connector to update all its metrics
    C::fetch_metrics(configuration, state).map_err(|_| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(models::ErrorResponse {
                message: "Unable to fetch metrics".into(),
                details: serde_json::Value::Null,
            }),
        )
    })?;

    let metric_families = metrics.gather();

    encoder.encode_to_string(&metric_families).map_err(|_| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(models::ErrorResponse {
                message: "Unable to encode metrics".into(),
                details: serde_json::Value::Null,
            }),
        )
    })
}

pub async fn get_capabilities<C: Connector>() -> Json<models::CapabilitiesResponse> {
    Json(C::get_capabilities().await)
}

pub async fn get_health<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
) -> Result<(), (StatusCode, Json<models::ErrorResponse>)> {
    // the context extractor will error if the deployment can't be found.
    match C::health_check(configuration, state).await {
        Ok(()) => Ok(()),
        Err(HealthError::Other(err)) => Err((
            StatusCode::SERVICE_UNAVAILABLE,
            Json(models::ErrorResponse {
                message: err.to_string(),
                details: serde_json::Value::Null,
            }),
        )),
    }
}

pub async fn get_schema<C: Connector>(
    configuration: &C::Configuration,
) -> Result<Json<models::SchemaResponse>, (StatusCode, Json<models::ErrorResponse>)> {
    Ok(Json(C::get_schema(configuration).await.map_err(
        |e| match e {
            crate::connector::SchemaError::Other(err) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(models::ErrorResponse {
                    message: "Internal error".into(),
                    details: serde_json::Value::Object(serde_json::Map::from_iter([(
                        "cause".into(),
                        serde_json::Value::String(err.to_string()),
                    )])),
                }),
            ),
        },
    )?))
}

pub async fn post_explain<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::QueryRequest>,
) -> Result<Json<models::ExplainResponse>, (StatusCode, Json<models::ErrorResponse>)> {
    Ok(Json(
        C::explain(configuration, state, request)
            .await
            .map_err(|e| match e {
                crate::connector::ExplainError::Other(err) => (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(models::ErrorResponse {
                        message: "Internal error".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "cause".into(),
                            serde_json::Value::String(err.to_string()),
                        )])),
                    }),
                ),
                crate::connector::ExplainError::InvalidRequest(detail) => (
                    StatusCode::BAD_REQUEST,
                    Json(models::ErrorResponse {
                        message: "Invalid request".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
                crate::connector::ExplainError::UnsupportedOperation(detail) => (
                    StatusCode::NOT_IMPLEMENTED,
                    Json(models::ErrorResponse {
                        message: "Unsupported operation".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
            })?,
    ))
}

pub async fn post_mutation<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::MutationRequest>,
) -> Result<Json<models::MutationResponse>, (StatusCode, Json<models::ErrorResponse>)> {
    Ok(Json(
        C::mutation(configuration, state, request)
            .await
            .map_err(|e| match e {
                crate::connector::MutationError::Other(err) => (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(models::ErrorResponse {
                        message: "Internal error".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "cause".into(),
                            serde_json::Value::String(err.to_string()),
                        )])),
                    }),
                ),
                crate::connector::MutationError::InvalidRequest(detail) => (
                    StatusCode::BAD_REQUEST,
                    Json(models::ErrorResponse {
                        message: "Invalid request".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
                crate::connector::MutationError::UnsupportedOperation(detail) => (
                    StatusCode::NOT_IMPLEMENTED,
                    Json(models::ErrorResponse {
                        message: "Unsupported operation".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
                crate::connector::MutationError::Conflict(detail) => (
                    StatusCode::CONFLICT,
                    Json(models::ErrorResponse {
                        message: "Request would create a conflicting state".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
                crate::connector::MutationError::ConstraintNotMet(detail) => (
                    StatusCode::FORBIDDEN,
                    Json(models::ErrorResponse {
                        message: "Constraint not met".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
            })?,
    ))
}

pub async fn post_query<C: Connector>(
    configuration: &C::Configuration,
    state: &C::State,
    Json(request): Json<models::QueryRequest>,
) -> Result<Json<models::QueryResponse>, (StatusCode, Json<models::ErrorResponse>)> {
    Ok(Json(
        C::query(configuration, state, request)
            .await
            .map_err(|e| match e {
                crate::connector::QueryError::Other(err) => (
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(models::ErrorResponse {
                        message: "Internal error".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "cause".into(),
                            serde_json::Value::String(err.to_string()),
                        )])),
                    }),
                ),
                crate::connector::QueryError::InvalidRequest(detail) => (
                    StatusCode::BAD_REQUEST,
                    Json(models::ErrorResponse {
                        message: "Invalid request".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
                crate::connector::QueryError::UnsupportedOperation(detail) => (
                    StatusCode::NOT_IMPLEMENTED,
                    Json(models::ErrorResponse {
                        message: "Unsupported operation".into(),
                        details: serde_json::Value::Object(serde_json::Map::from_iter([(
                            "detail".into(),
                            serde_json::Value::String(detail),
                        )])),
                    }),
                ),
            })?,
    ))
}

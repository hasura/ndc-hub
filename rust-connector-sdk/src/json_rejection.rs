//! We want errors returned from failed json extractors to be formatted as json as well.

use axum::extract;
use axum::{http::StatusCode, response::IntoResponse};
use ndc_client::models;

pub struct JsonRejection {
    error: models::ErrorResponse,
    status: StatusCode,
}

impl From<extract::rejection::JsonRejection> for JsonRejection {
    fn from(rejection: extract::rejection::JsonRejection) -> JsonRejection {
        JsonRejection {
            error: models::ErrorResponse {
                message: "Parse error".to_string(),
                details: serde_json::Value::String(rejection.body_text()),
            },
            status: rejection.status(),
        }
    }
}

impl IntoResponse for JsonRejection {
    fn into_response(self) -> axum::response::Response {
        let payload = serde_json::to_value(self.error).unwrap();
        (self.status, extract::Json(payload)).into_response()
    }
}

//! We want errors returned from failed json extractors to be formatted as json as well.

use axum::extract;
use axum::response::IntoResponse;
use ndc_client::models;

pub struct JsonRejection(extract::rejection::JsonRejection);

impl From<extract::rejection::JsonRejection> for JsonRejection {
    fn from(rejection: extract::rejection::JsonRejection) -> JsonRejection {
        JsonRejection(rejection)
    }
}

impl IntoResponse for JsonRejection {
    fn into_response(self: JsonRejection) -> axum::response::Response {
        let JsonRejection(rejection) = self;
        tracing::error!(
            meta.signal_type = "log",
            event.domain = "ndc",
            event.name = "Unable to deserialize request body",
            name = "Unable to deserialize request body",
            body = %rejection.body_text(),
            error = true,
        );
        let error = models::ErrorResponse {
            message: "Parse error".to_string(),
            details: serde_json::Value::String(rejection.body_text()),
        };
        let payload = serde_json::to_value(error).unwrap();
        (rejection.status(), extract::Json(payload)).into_response()
    }
}

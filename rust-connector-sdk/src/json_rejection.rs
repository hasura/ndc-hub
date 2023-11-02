//! We want errors returned from failed json extractors to be formatted as json as well.

use axum::extract;
use axum::{http::StatusCode, response::IntoResponse};

pub struct JsonRejection {
    message: String,
    status: StatusCode,
}

impl From<extract::rejection::JsonRejection> for JsonRejection {
    fn from(rejection: extract::rejection::JsonRejection) -> JsonRejection {
        JsonRejection {
            message: rejection.body_text(),
            status: rejection.status(),
        }
    }
}

impl IntoResponse for JsonRejection {
    fn into_response(self) -> axum::response::Response {
        let payload = serde_json::json!({
            "error": self.message,
        });

        (self.status, extract::Json(payload)).into_response()
    }
}

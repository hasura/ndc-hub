use axum::response::IntoResponse;
use bytes::Bytes;
use http::{header, HeaderValue};

/// Represents a response value that will be serialized to JSON.
///
/// The value may be of a type that implements `serde::Serialize`, or it may be
/// a contiguous sequence of bytes, which are _assumed_ to be valid JSON.
#[derive(Debug, Clone)]
pub enum JsonResponse<A> {
    /// A value that can be serialized to JSON.
    Value(A),
    /// A serialized JSON bytestring that is assumed to represent a value of
    /// type `A`. This is not guaranteed by the SDK; the connector is
    /// responsible for ensuring this.
    Serialized(Bytes),
}

impl<A> From<A> for JsonResponse<A> {
    fn from(value: A) -> Self {
        Self::Value(value)
    }
}

impl<A: (for<'de> serde::Deserialize<'de>)> JsonResponse<A> {
    /// Unwraps the value, deserializing if necessary.
    ///
    /// This is only intended for testing and compatibility. If it lives on a
    /// critical path, we recommend you avoid it.
    pub(crate) fn into_value<E: From<Box<dyn std::error::Error + Send + Sync>>>(
        self,
    ) -> Result<A, E> {
        match self {
            Self::Value(value) => Ok(value),
            Self::Serialized(bytes) => {
                serde_json::de::from_slice(&bytes).map_err(|err| E::from(Box::new(err)))
            }
        }
    }
}

impl<A: serde::Serialize> IntoResponse for JsonResponse<A> {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::Value(value) => axum::Json(value).into_response(),
            Self::Serialized(bytes) => (
                [(
                    header::CONTENT_TYPE,
                    HeaderValue::from_static(mime::APPLICATION_JSON.as_ref()),
                )],
                bytes,
            )
                .into_response(),
        }
    }
}

#[cfg(test)]
mod tests {
    use axum::{routing, Router};
    use axum_test_helper::TestClient;
    use reqwest::StatusCode;

    use super::*;

    #[tokio::test]
    async fn serializes_value_to_json() {
        let app = Router::new().route(
            "/",
            routing::get(|| async {
                JsonResponse::Value(Person {
                    name: "Alice Appleton".to_owned(),
                    age: 42,
                })
            }),
        );

        let client = TestClient::new(app);
        let response = client.get("/").send().await;

        assert_eq!(response.status(), StatusCode::OK);

        let headers = response.headers();
        assert_eq!(
            headers.get_all("Content-Type").iter().collect::<Vec<_>>(),
            vec!["application/json"]
        );

        let body = response.text().await;
        assert_eq!(body, r#"{"name":"Alice Appleton","age":42}"#);
    }

    #[tokio::test]
    async fn writes_json_string_directly() {
        let app = Router::new().route(
            "/",
            routing::get(|| async {
                JsonResponse::Serialized::<Person>(r#"{"name":"Bob Davis","age":7}"#.into())
            }),
        );

        let client = TestClient::new(app);
        let response = client.get("/").send().await;

        assert_eq!(response.status(), StatusCode::OK);

        let headers = response.headers();
        assert_eq!(
            headers.get_all("Content-Type").iter().collect::<Vec<_>>(),
            vec!["application/json"]
        );

        let body = response.text().await;
        assert_eq!(body, r#"{"name":"Bob Davis","age":7}"#);
    }

    #[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
    struct Person {
        name: String,
        age: u16,
    }
}

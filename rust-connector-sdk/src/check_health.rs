use std::net;

#[derive(Debug)]
pub enum HealthCheckError {
    RequestError(reqwest::Error),
    UnsuccessfulResponse {
        status: reqwest::StatusCode,
        body: String,
    },
}

impl std::fmt::Display for HealthCheckError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            HealthCheckError::RequestError(inner) => write!(f, "request error: {}", inner),
            HealthCheckError::UnsuccessfulResponse { status, body } => {
                write!(
                    f,
                    "unsuccessful response with status code: {}\nbody:\n{}",
                    status, body
                )
            }
        }
    }
}

impl std::error::Error for HealthCheckError {}

pub async fn check_health(socket: net::SocketAddr) -> Result<(), HealthCheckError> {
    let url = format!("http://{}/health", socket);
    let response = reqwest::get(url)
        .await
        .map_err(HealthCheckError::RequestError)?;
    let status = response.status();
    let body = response
        .text()
        .await
        .map_err(HealthCheckError::RequestError)?;
    if status.is_success() {
        Ok(())
    } else {
        Err(HealthCheckError::UnsuccessfulResponse { status, body })
    }
}

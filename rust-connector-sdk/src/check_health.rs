#[derive(Debug)]
pub enum HealthCheckError {
    ParseError(url::ParseError),
    RequestError(reqwest::Error),
    UnsuccessfulResponse {
        status: reqwest::StatusCode,
        body: String,
    },
}

impl std::fmt::Display for HealthCheckError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            HealthCheckError::ParseError(inner) => write!(f, "URL parse error: {}", inner),
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

pub async fn check_health(host: Option<String>, port: u16) -> Result<(), HealthCheckError> {
    let url = (|| -> Result<url::Url, url::ParseError> {
        let mut url = reqwest::Url::parse("http://localhost/").unwrap(); // cannot fail
        if let Some(host) = host {
            url.set_host(Some(&host))?;
        }
        url.set_port(Some(port)).unwrap(); // canont fail for HTTP URLs
        url.set_path("/health");
        Ok(url)
    })()
    .map_err(HealthCheckError::ParseError)?;
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

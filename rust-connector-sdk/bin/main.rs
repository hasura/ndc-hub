use std::process::ExitCode;

use ndc_sdk::connector::example::Example;
use ndc_sdk::default_main::default_main;

/// Run the [`Example`] connector using the [`default_main`]
/// function, which accepts standard configuration options
/// via the command line, configures metrics and trace
/// collection, and starts a server.
#[tokio::main]
pub async fn main() -> ExitCode {
    match default_main::<Example>().await {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("{}", err);
            ExitCode::FAILURE
        }
    }
}

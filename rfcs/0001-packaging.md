# Connector Packaging

## Purpose

For execution of queries and mutations, connectors are specified by the [NDC specification](http://hasura.github.io/ndc-spec/). However, for the purpose of deployment and configuration, their behavior is unspecified, or informally specified. 

This document exists to specify how connectors should be packaged in order to be accepted for inclusion in the Hasura Connector Hub. Any included connectors will be deployable via the CLI.

## Related Changes

_This RFC does not specify the following planned changes:_

- Work on existing connectors has shown that we need more configuration structure than a flat file. Therefore we plan to change the configuration file to a configuration directory with a supporting set of secrets in environment variables.
- There will also be no more `HasuraHubConnector` in v3-engine metadata. Instead the engine will only see connector URLs, and the CLI will manage the instantiation and deployment of connectors, and the creation of those URLs.

## Proposal

- A Hasura Hub data connector will be provided as a Docker image.
- The connector can expect configuration files to be mounted at `/etc/connector` inside the Docker image on startup. If the `HASURA_CONFIGURATION_DIRECTORY` environment variable is set, it should overwrite this default value.
  - The connector should not modify these files during execution, and can expect them not to be changed.
  - The `/etc/connector/.hasura` subdirectory (or `{HASURA_CONFIGURATION_DIRECTORY}/.hasura` in the general case) is reserved for future use and should not be used for configuration. Any connectors which enumerate all subdirectories of `/etc/connector`, for any reason, should ignore this subdirectory if it exists.
- The connector executable should accept the following subcommands:
  - `serve` should start a HTTP server on port `8080`, which is compatible with the NDC specification, with `/` as its base URL.
    - For example, `http://connector:8080/query` should implement the query endpoint
    - The default port can be overwritten using the `HASURA_CONNECTOR_PORT` environment variable.
- The image entrypoint should be set to the connector process, and the default `CMD` should be set to the `serve` command.
- The connector can read environment variables on startup for configuration purposes
  - The following environment variables are reserved, and should not be used for connector-specific configuration:
    - `HASURA_*`
  - Connectors can use environment variables as part of their configuration. Configuration that varies between different environments or regions (like connection strings) should be configurable via environment variables. 
- The connector should send any relevant trace spans in the OTLP format to the OTEL collector hosted at the URL provided by the `HASURA_OTLP_ENDPOINT` environment variable.
  - Spans should indicate the service name provided by the `HASURA_OTEL_SERVICE_NAME` environment variable.	
- If the `HASURA_SERVICE_TOKEN_SECRET` environment variable is specified and non-empty, then the connector should implement bearer-token HTTP authorization using the provided static secret token.
- Information log messages should be logged in plain text to standard output.
- Error messages should be logged in plain text to standard error.
- On startup, in case of failure to start, the connector should flush any error messages to standard error, and exit with a non-zero exit code.
- The connector should respond to the following signals:
  - `SIGTERM`/`SIGINT` - gracefully shutdown the server, and stop the connector process
- The connector should start as quickly as possible, without any build steps, by reading configuration from disk. Build steps should be performed in the construction of the Docker image, not on startup.
  - If there is a build step which depends on files controlled by the user (for example, installing dependencies), then they should be moved into a supporting `Dockerfile`. The Docker image which represents the connector would then be the result of `docker build` on this `Dockerfile`. 
    - To support this use case, tooling should be provided to build a Docker image from a local `Dockerfile`, which may refer to existing Hub connector images.

## Open Questions

- Do we need a `validate` subcommand to support the LSP/CLI?
- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?
# Connector Deployment

## Purpose

For execution of queries and mutations, connectors are specified by the [NDC specification](http://hasura.github.io/ndc-spec/). However, for the purpose of deployment and configuration, their behavior is unspecified, or informally specified. 

This document exists to specify how connectors should be packaged in order to be accepted for inclusion in the Hasura Connector Hub. Any included connectors will be deployable via the CLI.

### Out of Scope

This RFC does not concern itself with the DX aspects of connector metadata authoring, development etc. in the CLI. As it relates to the connector hub, those aspects will be specified in a separate RFC.

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
- The image `ENTRYPOINT` should be set to the connector process, and the default `CMD` should be set to the `serve` command. This can mean setting it to the connector executable itself, or some intermediate executable/script that eventually provides its command line arguments to the connector executable.
- The connector can read environment variables on startup for configuration purposes
  - The following environment variables are reserved, and should not be used for connector-specific configuration:
    - `HASURA_*`
    - `OTEL_EXPORTER_*`
  - Connectors can use environment variables as part of their configuration. Configuration that varies between different environments or regions (like connection strings) should be configurable via environment variables. 
- The connector should send any relevant trace spans in the OTLP format to the OTEL collector hosted at the URL provided by the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable.
  - Spans should indicate the service name provided by the `OTEL_SERVICE_NAME` environment variable.	
- If the `HASURA_SERVICE_TOKEN_SECRET` environment variable is specified and non-empty, then the connector should implement bearer-token HTTP authorization using the provided static secret token.
- Information log messages should be logged in plain text to standard output.
- Error messages should be logged in plain text to standard error.
- On startup, in case of failure to start, the connector should flush any error messages to standard error, and exit with a non-zero exit code.
- The connector should respond to the following signals:
  - `SIGTERM`/`SIGINT` - gracefully shutdown the server, and stop the connector process
- The connector should start as quickly as possible, without any build steps, by reading configuration from disk. Build steps should be performed in the construction of the Docker image, not on startup.
  - To support these build steps, tooling should support building images from Dockerfiles. See "Deployment API" below.

### Open Questions

- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?

## Deployment API

Docker images should be built in the same environment as which they are run, to avoid possible issues with differences in architecture, etc. Therefore, we need to specify _how to build_ images which meet the specification above. 

The _connector build request_ data structure describes the ways in which we can build such an image unambiguously. There are two alternatives: from a named and versioned hub connector, or from a Dockerfile.

Here is a sketch of the data structure in Rust:

```rust
pub enum ConnectorBuildRequest {
  FromHubConnector {
    name: String,
    version: Version, // sha hash
  },
  FromDockerfileAndBuildInputs {
    tar: TarFile,
  }
}
```

How this structure gets built by CLI (or its supporting web service) is out of scope. For example, we might fetch tar bundles from Git repos, or from the filesystem. Dockerfiles might be under the user's control, or not. But this structure is what is required to build images for deployment.

In the case of `FromHubConnector`, the expectation is that the connector build service maintains a list of prebuilt images, indexed by the names and versions of hub connectors.

Here, in the case of `FromHubConnector`, a full directory containing a `Dockerfile` and any supporting build inputs is provided as the bytes of a `.tar` file, but the exact protocol can be up to the service implementer.
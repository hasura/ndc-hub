# Connector Packaging

## Purpose

For execution of queries and mutations, connectors are specified by the [NDC specification](http://hasura.github.io/ndc-spec/). However, for the purpose of deployment and configuration, their behavior is unspecified, or informally specified. 

This document exists to specify how connectors should be packaged in order to be accepted for inclusion in the Hasura Connector Hub. Any included connectors will be deployable via the CLI.

## Related Changes

_This RFC does not specify the following planned changes:_

- Work on existing connectors has shown that we need more configuration structure than a flat file. Therefore we plan to change the configuration file to a configuration directory with a supporting set of secrets in environment variables.
- There will also be no more `HasuraHubConnector` in v3-engine metadata. Instead the engine will only see connector URLs, and the CLI will manage the instantiation and deployment of connectors, and the creation of those URLs.

## Proposal
Hasura Hub data connectors are packaged as Docker images which expect to receive a specific `CMD` format and expect specific Hasura-defined environment variables to be specified in order to control common connector configuration.

### Packaging Categories
There are two different categories of Docker-based packaging.

#### Connectors that do not require a build step
These connectors have a pre-built Docker image that simply requires the configuration files to be provided at runtime to execute correctly. They do not require a time-consuming build step to occur based on the configuration files before the connector can run, and therefore can start immediately.

- The connector can expect configuration files to be mounted at `/etc/connector` inside the Docker image on startup. This may be done via volume mounting, or it may be done by including the files in the docker image itself. 

- If the `HASURA_CONFIGURATION_DIRECTORY` environment variable is set, the connector should look for the configuration files in the location specified by the environment variable.

- The connector should not modify these files during execution, and can expect them not to be changed.

- The `/etc/connector/.hasura` subdirectory (or `{HASURA_CONFIGURATION_DIRECTORY}/.hasura` in the general case) is reserved for future use and should not be used for configuration. Any connectors which enumerate all subdirectories of `/etc/connector`, for any reason, should ignore this subdirectory if it exists.

#### Connectors that require a build step
These connectors require a time-consuming build step to be performed over their configuration before they can execute correctly. Once the build step is performed, the results can be retained, and they can start immediately.

- The build steps are defined in a supporting Dockerfile (for example, installing dependencies). The configuration files (which are build inputs) will be provided to the Dockerfile as its build context.

- The result of the Dockerfile build should be a connector Docker image that contains the configuration files and can start immediately.

### Common Docker Image Behaviour
The Docker images of both categories of connectors behave in the same way with respect to the following points:

- The connector executable should accept the following subcommands:
  - `serve` should start a HTTP server on port `8080`, which is compatible with the NDC specification, with `/` as its base URL.
    - For example, `http://connector:8080/query` should implement the query endpoint
    - The default port can be overwritten using the `HASURA_CONNECTOR_PORT` environment variable.

- The image `ENTRYPOINT` should be set to the connector process, and the default `CMD` should be set to the `serve` command. This can mean setting it to the connector executable itself, or some intermediate executable/script that eventually provides its command line arguments to the connector executable.

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

### Watch Mode
When developing locally using the connector, the user may utilize the connector in a "watch mode" that automatically reloads the connector with new configuration as they change the configuration.

#### Connectors that do not require a build step
If any configuration file changes, the connector image is restarted with the new version of the configuration file mounted.

#### Connectors that require a build step
These connectors can provide a [docker compose watch configuration](https://docs.docker.com/compose/compose-file/develop/#watch) that defines how to watch the configuration files and how to react to different files changing. It can react by causing a container rebuild and restart, or by copying the new file inside the existing container and optionally restarting the container. This allows the connector to optimise its hot reload capability by only performing a rebuild where necessary.

For example, a NodeJS-based connector that needs to perform an npm package restore as a part of its Docker build may use the following docker compose watch configuration:

```yaml
# Rebuild the container if a new package restore is required because package[-lock].json changed
- path: package.json
  target: /etc/connector/package.json
  action: rebuild
- path: package-lock.json
  target: /etc/connector/package-lock.json
  action: rebuild
# For any other file change, simply copy it into the existing container and restart it
- path: ./
  target: /etc/connector
  action: sync+restart
```

## Open Questions
- Where does the connector declare what type of packaging it uses and its watch configuration?
- Do we need a `validate` subcommand to support the LSP/CLI?
- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?
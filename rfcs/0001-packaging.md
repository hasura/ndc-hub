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

They also define a Hasura Hub Connector Definition that Hasura tooling uses to understand how to interact with the connector and its packaging.

### Docker Packaging Categories
There are two different categories of Docker-based packaging.

#### Connectors that do not require a build step (`PrebuiltDockerImagePackaging`)
These connectors have a pre-built Docker image that simply requires the configuration files to be provided at runtime to execute correctly. They do not require a time-consuming build step to occur based on the configuration files before the connector can run, and therefore can start immediately.

- The connector can expect configuration files to be mounted at `/etc/connector` inside the Docker image on startup. This may be done via volume mounting, or it may be done by including the files in the docker image itself. 

- If the `HASURA_CONFIGURATION_DIRECTORY` environment variable is set, the connector should look for the configuration files in the location specified by the environment variable.

- The connector should not modify these files during execution, and can expect them not to be changed.

- The `/etc/connector/.hasura` subdirectory (or `{HASURA_CONFIGURATION_DIRECTORY}/.hasura` in the general case) is reserved for future use and should not be used for configuration. Any connectors which enumerate all subdirectories of `/etc/connector`, for any reason, should ignore this subdirectory if it exists.

#### Connectors that require a build step (`DockerBuildPackaging`)
These connectors require a time-consuming build step to be performed over their configuration before they can execute correctly. Once the build step is performed, the results can be retained, and they can start immediately.

- The build steps are defined in a supporting Dockerfile (for example, installing dependencies). The configuration files (which are build inputs) will be provided to the Dockerfile as its build context.

- The result of the Dockerfile build should be a connector Docker image that contains the configuration files and can start immediately.

### Common Docker Image Behaviour
The Docker images of both categories of connectors behave in the same way with respect to the following points:

- The connector executable should accept the following subcommands:
  - `serve` should start a HTTP server on port `8080`, which is compatible with the NDC specification, with `/` as its base URL.
    - For example, `http://connector:8080/query` should implement the query endpoint
    - The default port can be overwritten using the `PORT` environment variable.

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

### Hasura Hub Connector Definition
In order for the Hasura tooling to understand a connector, know how to interact with it and know how it is packaged, a connector will need a definition that contains this information.

The connector definition package is a .tgz (tar file, gzipped) that contains the following files:

```
/
  .hasura/
    connector-metadata.json # See ConnectorMetadataDefinition type
    docker-compose.yaml # If DockerComposeWatchMode is used
    
    ### Start: If Docker packaging requires a build step
    .dockerignore
    Dockerfile
    ### End: If Docker packaging requires a build step
    
  ### Start: Connector specific configuration files
  src/
    functions.ts
  package.json
  package-lock.json
  tsconfig.json
  ### End: Connector specific configuration files

```

```typescript
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
}

type PackagingDefinition = PrebuiltDockerImagePackaging | DockerBuildPackaging

type PrebuiltDockerImagePackaging = {
  type: "PrebuiltDockerImage"
  dockerImage: string // eg "hasura/postgres-data-connector:1.0.0"
}

type DockerBuildPackaging = {
  type: "DockerBuild"
  watchMode: WatchMode
}

type WatchMode = DockerRebuildWatchMode | DockerComposeWatchMode | ShellCommandWatchMode

type DockerRebuildWatchMode = {
  type: "DockerRebuild"
}

type DockerComposeWatchMode = {
  type: "DockerCompose"
}

type ShellCommandWatchMode = {
  type: "ShellCommand"
  command: string
}

type EnvironmentVariableDefinition = {
  name: string
  description: string
  defaultValue?: string
}
```

The `.hasura/connector-metadata.json` contains JSON that describes:
- The environment variables the connector supports to configure it (`supportedEnvironmentVariables`)
- The packaging definition, which can be either `PrebuiltDockerImagePackaging` (a connector that does not require a build step), or  `DockerBuildPackaging` (a connector that requires a build step).
  - `PrebuiltDockerImagePackaging` defines the prebuilt `dockerImage` used to run the connector (`dockerImage`)
  - If `DockerBuildPackaging` is used, a Dockerfile must be in the `.hasura` directory (and optionally, a `.dockerignore`). It will be used to build the connector. A `watchMode` must be specified, to tell the Hasura tooling how to perform watching for this connector (see [Watch Mode section](#watch-mode)).

#### Watch Mode
When developing locally using the connector, the user may utilize the connector in a "watch mode" that automatically reloads the connector with new configuration as they change the configuration.

#### Connectors that do not require a build step (`PrebuiltDockerImagePackaging`)
If any configuration file changes, the connector image is restarted with the new version of the configuration file mounted.

#### Connectors that require a build step (`DockerBuildPackaging`)
There are three options a connector can choose from in their `connector-metadata.json`:

##### `DockerRebuildWatchMode`
In this mode, if any configuration file changes, the Docker container is rebuilt and restarted. This is the simplest mode, but the least performant, unless your Docker builds are extremely quick.

##### `DockerComposeWatchMode`
In this mode, connectors can provide a `.hasura/docker-compose.yaml` file. This file should:

- Define a single service, which represents the connector 
- Build that service using the `.hasura/Dockerfile`
- Define a port mapping that uses the environment variable `PORT` for the host port and 8080 for the container port. This allows the watch tooling to control the port used by the connector and exposes the port on the host.
- Provide a [docker compose watch configuration](https://docs.docker.com/compose/compose-file/develop/#watch) that defines how to watch the configuration files and how to react to different files changing. It can react by causing a container rebuild and restart, or by copying the new file inside the existing container and optionally restarting the container. This allows the connector to optimise its hot reload capability by only performing a rebuild where necessary.

For example, a NodeJS-based connector that needs to perform an npm package restore as a part of its Docker build may define the following `.hasura/docker.compose.yaml`:

```yaml
services:
  connector:
    build: 
      context: ../ # The build context is the parent directory of .hasura/
      dockerfile: ./Dockerfile
    ports:
      - ${PORT}:8080 # Required so that the watch tooling can control the port used
    develop:
      watch:
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

##### `ShellCommandWatchMode`
In this mode, connectors can provide some native way of performing a hot-reloading watch mode. Whatever this custom method is, it will be started by running the shell command defined. This can be useful to allow local tooling to be used to run the connector and perform hot reloading.

* The shell command will be executed with the working directory set to the root directory where the configuration files are located.
* The PORT environment variable will be set and must be respected. The connector must start serving on this port.
* SIGINT and SIGSTOP signals must be respected and must cause the watch mode and connector to shut down
* Any stdout and stderr output will be collected by the Hasura tooling for display

For example, for the NodeJS Lambda Connector, it could set the watch shell command to `npm run watch`, which would run the connector and activate its built-in hot-reloading functionality.

## Open Questions
### Custom CLI Plugins
- Do connectors need to declare if they have a custom CLI plugin?
- If custom CLI plugins provide configuration update services (such as DB schema introspection), does this need to be integrated with watch mode, and if so, how is this declared?
- Do we need a `validate` subcommand to support the LSP/CLI?

### Publishing the Hasura Hub Connector Definition
- How is this published to the Hub?
  
### OpenTelemetry
- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?
# Connector Packaging

## Purpose

Connector behavior is specified by the [NDC specification](http://hasura.github.io/ndc-spec/), and the deployment API is specified in the [deployment RFC](./0000-deployment.md). This document specifies how the Hasura CLI and Hasura Connector Hub can work together to facilitate connector development, packaging and deployment via the API.

## Proposal

### Docker Packaging Categories

Hasura Hub data connectors are packaged as Docker images which follow the [deployment specification](./0000-deployment.md). Per the deployment specification, there are two ways in which the CLI can describe those images to the deployment service:

- Connectors that do not require a build step can be described using a name and version, which should correspond to an entry in the connector hub registry.
- Connectors that require a build step (performed over some build inputs before they can execute correctly) can be described using a Dockerfile bundled with any additional build inputs.

### Connector Configuration

_TODO_: describe the layout of files on disk for prebuilt connectors, and for connectors which require a build step. Is this the same as the layout described in "Hasura Hub Connector Definition" below?

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
### Build Inputs vs Configuration
- Is there a difference between Docker build inputs and connector configuration files (currently the RFC does not distinguish these)? If so:
  - Can connector configuration be optional (defaulted to zero files) where it is not used (ie NodeJS Lambda Connector)
  - How is the difference represented on disk in the user's Hasura project? (Different directories?)

### Custom CLI Plugins
- Do connectors need to declare if they have a custom CLI plugin?
- If custom CLI plugins provide configuration update services (such as DB schema introspection), does this need to be integrated with watch mode, and if so, how is this declared?
- Do we need a `validate` subcommand to support the LSP/CLI?

### Publishing the Hasura Hub Connector Definition
- How is this published to the Hub?
  
### OpenTelemetry
- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?
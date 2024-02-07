# Connector Packaging and Hub Connector Definition

## Purpose

Connector behavior is specified by the [NDC specification](http://hasura.github.io/ndc-spec/), and the deployment API is specified in the [deployment RFC](./0000-deployment.md). This document specifies how the Hasura CLI and Hasura Connector Hub can work together to facilitate connector development, packaging and deployment via the API.

## Proposal

### Docker Packaging Categories

Hasura Hub data connectors are packaged as Docker images which follow the [deployment specification](./0000-deployment.md). Per the deployment specification, there are two ways in which the CLI can describe those images to the deployment service:

- Connectors that do not require a build step can be described using a name and version, which should correspond to an entry in the connector hub registry.
- Connectors that require a build step (performed over some build inputs before they can execute correctly) can be described using a Dockerfile bundled with any additional build inputs.

### Hasura Hub Connector Definition
In order for the Hasura tooling to understand a connector, know how to interact with it and know how it is packaged, a connector will need a definition that contains this information.

The connector definition is a file structure that contains the following files and directories:

```
/
  connector-metadata.json # See ConnectorMetadataDefinition type
  
  docker-compose.yaml

  build-files/ # Connector specific build/configuration files
    src/
      functions.ts
    package.json
    package-lock.json
    tsconfig.json

  docker/ # If ManagedDockerBuildPackaging is used
    .dockerignore
    Dockerfile
```

```typescript
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
  cliPlugin?: CliPluginDefinition
}

type PackagingDefinition = PrebuiltDockerImagePackaging | ManagedDockerBuildPackaging

type PrebuiltDockerImagePackaging = {
  type: "PrebuiltDockerImage"
  dockerImage: string // eg "hasura/postgres-data-connector:1.0.0"
}

type ManagedDockerBuildPackaging = {
  type: "ManagedDockerBuild"
  additionalWatchModes: ShellCommandWatchMode[]
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

type CliPluginDefinition = {
  pluginPackageUrl: string // https://url/to/download/plugin/to/install.tgz
  pluginPackageSHA256: string // To ensure plugin integrity and also to identify which version to use out of those installed
  watchSubcommand?: string // If the connector wants to add to the watch mode, it can define what subcommand to use "watch --whatever --flags"
}
```

The `connector-metadata.json` contains JSON that describes:
- The environment variables the connector supports to configure it (`supportedEnvironmentVariables`)
- The packaging definition, which can be either `PrebuiltDockerImagePackaging` (a connector that does not require a build step), or  `ManagedDockerBuildPackaging` (a connector that requires a build step).
  - `PrebuiltDockerImagePackaging` defines the prebuilt `dockerImage` used to run the connector (`dockerImage`)
  - If `ManagedDockerBuildPackaging` is used, a Dockerfile must be in the `.hasura` directory (and optionally, a `.dockerignore`). It will be used to build the connector.
- An optional `CLIPluginDefinition` that describes where to acquire the CLI plugin for this connector that can be used to enhance watch mode with connector-specific configuration updates

### Watch Mode
When developing locally using the connector, the user may utilize the connector in a "watch mode" that automatically reloads the connector with new configuration as they change the configuration.

Every connector must specify a `docker-compose.yaml` in their connector definition that defines how to watch the container.

#### Connectors that do not require a build step (`PrebuiltDockerImagePackaging`)
The `docker-compose.yaml` should contain something like this:

```yaml
services:
  connector:
    develop:
      watch:
        - path: ./
          target: /etc/connector
          action: sync+restart
```

The `connector` service must be defined, and note that no `image` or `build` property is defined. The CLI tooling will [extend](https://docs.docker.com/compose/compose-file/05-services/#extends) from this to add the necessary pre-built image. This exists solely to provide the watch configuration.

#### Connectors that require a build step (`ManagedDockerBuildPackaging`)
The `docker-compose.yaml` should contain something like this (specific to the watch requirements of the connector):

```yaml
services:
  connector:
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

The `connector` service must be defined, and note that no `image` or `build` property is defined. The CLI tooling will [extend](https://docs.docker.com/compose/compose-file/05-services/#extends) from this to add the necessary `build` property based on the build context from `ConnectorManifest` and the Dockerfile included in the Connector Definition. This exists solely to provide the watch configuration.

`ManagedDockerBuildPackaging` connectors can also optionally provide additional watch modes. This is intended to support connectors that can run using local developer machine tooling (for example, local NodeJS for debugging purposes). These additional watch modes may be chosen by the user to run _instead_ of the docker compose watch-based watch mode. They are configured via the `additionalWatchModes` property, which currently only supports `ShellCommandWatchMode`.

##### `ShellCommandWatchMode`
In this mode, connectors can provide some native way of performing a hot-reloading watch mode. Whatever this custom method is, it will be started by running the shell command defined. This can be useful to allow local tooling to be used to run the connector and perform hot reloading.

* The shell command will be executed with the working directory set to the root directory where the configuration files are located.
* The PORT environment variable will be set and must be respected. The connector must start serving on this port.
* SIGINT and SIGSTOP signals must be respected and must cause the watch mode and connector to shut down
* Any stdout and stderr output will be collected by the Hasura tooling for display

For example, for the NodeJS Lambda Connector, it could set the watch shell command to `npm run watch`, which would run the connector and activate its built-in hot-reloading functionality.

#### CLI Plugin Watch Mode
If connectors need to extend the watch process with their own functionality, they can implement a CLI plugin that contains a `watch` subcommand. If specified in the `connector-metadata.json` the tooling will use it in addition to existing watch functionality.

An example use case for this would be for a Postgres connector, where you may want to watch the database for schema changes, and when they occur, update the schema introspection configuration file in the connector's build directory with the latest schema details. The docker compose watch would then reload the connector with the updated introspection configuration file.

### Connector Layout in Hasura Project
When a new connector is added to a Hasura project using `hasura3 add ConnectorManifest --hub hasura/nodejs-lambda:1.0`, the CLI acquires the Hasura Hub Connector Definition for the specified Hub Connector. It then places this inside the `~/.hasura/hub-connectors/` directory. This is done to keep non-user editable files out of the user's source tree.

The CLI then puts the `build-files/` into the build directory for that connector and adds a `.build.hml` file with the `ConnectorManifest` metadata object.

```
# A copy of the `Hasura Hub Connector Definition` for a particular hub connector version
~/.hasura/hub-connectors/hasura/nodejs-lambda/1.0/ 
  connector-metadata.json
  docker-compose.yaml
  build-files/
    src/
      functions.ts
    package.json
    package-lock.json
    tsconfig.json
  docker/
    .dockerignore
    Dockerfile

project/
  subgraphs/default/dataconnectors/
    my-dataconnector/
      build/
        ### Start: from `build-files/`
        src/
          functions.ts
        package.json
        package-lock.json
        tsconfig.json
        ### End: from `build-files/`

        my-dataconnector.build.hml # Connector manifest, see below
```

Here's an example of what `my-dataconnector.build.hml` would contain (`ConnectorManifest` is not specified by this RFC):
```yaml
kind: ConnectorManifest
definition:
  name: my-dataconnector
  type: local
  hub-connector: hasura/nodejs-lambda:1.0 # This is used to find the Hasura Hub Connector Definition in ~/.hasura/hub-connectors/
  tunnel: true
  instances:
  - build:
      context: .
      env:
        STRIPE_API_ENDPOINT: https://api.stripe.com/
```

## Open Questions
### Build Inputs vs Configuration
- Is there a difference between Docker build inputs and connector configuration files (currently the RFC does not distinguish these)? If so:
  - Can connector configuration be optional (defaulted to zero files) where it is not used (ie NodeJS Lambda Connector)
  - How is the difference represented on disk in the user's Hasura project? (Different directories?)

### Custom CLI Plugins
- `watchSubcommand` seems redundant if we're going to have a CLI contract. We could just have `watch: boolean` instead. Or perhaps the CLI plugin can declare what it supports and this exists outside of the connector-metadata.json?

### Publishing the Hasura Hub Connector Definition
- How is this published to the Hub? Git? `tar.gz`?
  
### OpenTelemetry
- Do we want to reserve environment variables `OTEL_*` for possible future use of the [OTLP exporter spec](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md)?
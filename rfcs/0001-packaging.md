# Connector Packaging and Hub Connector Definition

## Purpose

Connector behavior is specified by the [NDC specification](http://hasura.github.io/ndc-spec/), and the deployment API is specified in the [deployment RFC](./0000-deployment.md). This document specifies how the Hasura CLI and Hasura Connector Hub can work together to facilitate connector development, packaging and deployment via the API.

## Proposal

### Docker Packaging Categories

Hasura Hub data connectors are packaged as Docker images which follow the [deployment specification](./0000-deployment.md). Per the deployment specification, there are two ways in which the CLI can describe those images to the deployment service:

- Connectors that do not require a build step can be described using a name and version, which should correspond to an entry in the connector hub registry.
- Connectors that require a build step (performed over some build inputs before they can execute correctly) can be described using a Dockerfile bundled with any additional build inputs.

### Connector Definition
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
}

type EnvironmentVariableDefinition = {
  name: string
  description: string
  defaultValue?: string
}

// This is subject to change based on CLI plugin spec discussions
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

#### CLI Plugin Watch Mode
If connectors need to extend the watch process with their own functionality, they can implement a CLI plugin that contains a `watch` subcommand. If specified in the `connector-metadata.json` the tooling will use it in addition to existing watch functionality.

An example use case for this would be for a Postgres connector, where you may want to watch the database for schema changes, and when they occur, update the schema introspection configuration file in the connector's build directory with the latest schema details. The docker compose watch would then reload the connector with the updated introspection configuration file.

Note that the exact definition of how this needs to be specified in the Connector Definition will need to wait until how CLI plugins work is specified.

### Connector Layout in Hasura Project
When a new connector is added to a Hasura project using `hasura3 add ConnectorManifest --hub hasura/nodejs-lambda:1.0`, the CLI acquires the Connector Definition for the specified Hub Connector. It then places this inside the `~/.hasura/hub-connectors/` directory. This is done to keep non-user editable files out of the user's source tree.

The CLI then puts the `build-files/` into the build directory for that connector and adds a `.build.hml` file with the `ConnectorManifest` metadata object.

```
# A copy of the `Connector Definition` for a particular hub connector version
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
        ### Start: copied from `build-files/`
        src/
          functions.ts
        package.json
        package-lock.json
        tsconfig.json
        ### End: copied from `build-files/`

        my-dataconnector.build.hml # Connector manifest, see below
```

Here's an example of what `my-dataconnector.build.hml` would contain (`ConnectorManifest` is not specified by this RFC):
```yaml
kind: ConnectorManifest
definition:
  name: my-dataconnector
  type: local
  connector:
    type: hub
    name: hasura/nodejs-lambda:1.0 # This is used to find the Hasura Hub Connector Definition in ~/.hasura/hub-connectors/
  tunnel: true
  instances:
  - build:
      context: .
      env:
        STRIPE_API_ENDPOINT: https://api.stripe.com/
```

It is worth clarifying that the contents what of the `build` directory in the Hasura project source tree contains changes depending on whether we are building a `PrebuiltDockerImagePackaging` connector versus a `ManagedDockerBuildPackaging` connector. For `PrebuiltDockerImagePackaging`, the build directory contains the configuration files that will be mounted to the `/etc/connector` directory of the Docker image at runtime. For `ManagedDockerBuildPackaging`, the build directory contains the Docker build context that is handed to the Dockerfile when it is built. There are no configuration files to be mounted at `/etc/connector` for `ManagedDockerBuildPackaging` connectors, as it is assumed that any relevant configuration will be built into the container image by the Dockerfile. In practice this means an empty volume will be mounted `/etc/connector` at runtime.

#### Non-Hasura Hub Connector
To support the use case of a connector developer being and to share/iterate on a connector and use it inside a Hasura project without first publishing it to the Hasura Connector Hub, we will support the ability to inline a Connector Definition inside the Hasura project source tree. This enables the Hasura tooling to still know how to interact with the connector, but does not require it to be published on the Connector Hub.

The CLI could support the use of non-Hasura Hub Connectors by supporting the use of `.tgz` archives of the Connector Definition: `hasura3 add ConnectorManifest --inline my-connector-definition.tgz`

```
project/
  subgraphs/default/dataconnectors/
    my-dataconnector/
      connector-definition/ # `my-connector-definition.tgz` extracted into the source tree
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
      build/
        ### Start: copied from `build-files/`
        src/
          functions.ts
        package.json
        package-lock.json
        tsconfig.json
        ### End: copied from `build-files/`

        my-dataconnector.build.hml # Connector manifest, see below
```

```yaml
kind: ConnectorManifest
definition:
  name: my-dataconnector
  type: local
  connector:
    type: inline # Tells the tooling to look for the connector definition in `connector-definition/`
  tunnel: true
  instances:
  - build:
      context: .
      env:
        STRIPE_API_ENDPOINT: https://api.stripe.com/
```

## Out of Scope for this RFC
* How the Connector Definition is published to the Hasura Connector Hub (can assume for starters that the Connector Definition will be tar-gzipped into an archive and submitted somewhere)
* How CLI plugins work

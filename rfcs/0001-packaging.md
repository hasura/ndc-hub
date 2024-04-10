# Connector Packaging and Hub Connector Definition

> [!NOTE]
> This RFC has since been extended by the [Native Packaging RFC](./0004-native-packaging.md)

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
  .hasura-connector/
    connector-metadata.yaml # See ConnectorMetadataDefinition type
    ### Start: If ManagedDockerBuildPackaging is used
    .dockerignore
    Dockerfile
    ### End: If ManagedDockerBuildPackaging is used

  ### Start: Connector specific build/configuration files
  src/
    functions.ts
  package.json
  package-lock.json
  tsconfig.json
  ### End: Connector specific build/configuration files

```

```typescript
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
  commands: Commands
  cliPlugin?: CliPluginDefinition
  dockerComposeWatch: DockerComposeWatch
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

type Commands = {
  update?: string
  watch?: string
}

type CliPluginDefinition = {
  name: string
  version: string
}

// From: https://github.com/compose-spec/compose-spec/blob/1938efd103f8e0817ca90e5f15177ec0317bbaf8/schema/compose-spec.json#L455
type DockerComposeWatch = DockerComposeWatchItem[]

type DockerComposeWatchItem = {
  path: string
  action: "rebuild" | "sync" | "sync+restart"
  target?: string
  ignore?: string[]
}

```

The `connector-metadata.yaml` contains YAML that describes:
- The environment variables the connector supports to configure it (`supportedEnvironmentVariables`)
- The packaging definition, which can be either `PrebuiltDockerImagePackaging` (a connector that does not require a build step), or  `ManagedDockerBuildPackaging` (a connector that requires a build step).
  - `PrebuiltDockerImagePackaging` defines the prebuilt `dockerImage` used to run the connector (`dockerImage`)
  - If `ManagedDockerBuildPackaging` is used, a Dockerfile must be in the `.hasura3` directory (and optionally, a `.dockerignore`). It will be used to build the connector.
- A `commands` structure that optionally defines what shell commands to run for an "update" (eg. refresh schema introspection details) and "watch" (eh. watch and refresh schema introspection details periodically) scenario.
- An optional `CLIPluginDefinition` that describes where to acquire the CLI plugin for this connector that can be used by the `commands` structure. If provided, the CLI plugin executable will be made available on the `PATH` for the commands and some configuration environment variables will be set (see the [CLI plugin RFC](https://github.com/hasura/ndc-hub/blob/cli-guidelines/rfcs/0002-cli-guidelines.md) for more details).
- A `dockerComposeWatch` that defines how to rebuild/restart the container if the user modifies their connector configuration in their project (see below)

### Watch Mode
When developing locally using the connector, the user may utilize the connector in a "watch mode" that automatically reloads the connector with new configuration as they change the configuration.

Every connector must specify a `dockerComposeWatch` property in their `connector-metadata.yaml` that defines how to watch the container.

#### Connectors that do not require a build step (`PrebuiltDockerImagePackaging`)
The `connector-metadata.yaml` should contain something like this:

```yaml
dockerComposeWatch:
  - path: ./
    target: /etc/connector
    action: sync+restart
```

When configuration files change, this simply copies them into the container and restarts it.

#### Connectors that require a build step (`ManagedDockerBuildPackaging`)
The `connector-metadata.yaml` should contain something like this (specific to the watch requirements of the connector):

```yaml
dockerComposeWatch:
  # Rebuild the container if a new package restore is required because package[-lock].json changed
  - path: package.json
    target: /functions/package.json
    action: rebuild
  - path: package-lock.json
    target: /functions/package-lock.json
    action: rebuild
  # For any other file change, simply copy it into the existing container and restart it
  - path: ./
    target: /functions
    action: sync+restart
```

When the package.json or package-lock.json files change, the container is rebuilt, but for all other files, they are simply copied into the container and it is restarted.

#### CLI Plugin Watch Mode
If connectors need to extend the watch process with their own functionality, they can implement a CLI plugin that contains a `watch` subcommand. If specified in the `connector-metadata.yaml` the tooling will use it in addition to existing watch functionality.

An example use case for this would be for a Postgres connector, where you may want to watch the database for schema changes, and when they occur, update the schema introspection configuration file in the connector's build directory with the latest schema details. The docker compose watch would then reload the connector with the updated introspection configuration file.

Note that the exact definition of how this needs to be specified in the Connector Definition will need to wait until how CLI plugins work is specified.

### Connector Layout in Hasura Project
When a new connector is added to a Hasura project using `hasura3 add ConnectorManifest --hub hasura/nodejs-lambda:1.0`, the CLI acquires the Connector Definition for the specified Hub Connector. It then places this inside the `~/.hasura3/hub-connectors/` directory. This is done to keep non-user editable files out of the user's source tree.

The CLI then puts the `build-files/` into the build directory for that connector and adds a `.build.hml` file with the `ConnectorManifest` metadata object.

```
# A copy of the `Connector Definition` for a particular hub connector version
~/.hasura3/hub-connectors/hasura/nodejs-lambda/1.0/
  .hasura-connector/
    connector-metadata.yaml
    .dockerignore
    Dockerfile
  src/
    functions.ts
  package.json
  package-lock.json
  tsconfig.json

project/
  subgraphs/default/dataconnectors/
    my-dataconnector/
      build/
        ### Start: Unpack `Connector Definition` and delete `.hasura-connector/`
        src/
          functions.ts
        package.json
        package-lock.json
        tsconfig.json
        ### End: Unpack `Connector Definition` and delete `.hasura-connector/`

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
    name: hasura/nodejs-lambda:1.0 # This is used to find the Hasura Hub Connector Definition in ~/.hasura3/hub-connectors/
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
      build/
        ### Start: `my-connector-definition.tgz` extracted into the source tree
        .hasura-connector/
          connector-metadata.yaml
          .dockerignore
          Dockerfile
        src/
          functions.ts
        package.json
        package-lock.json
        tsconfig.json
        ### End: `my-connector-definition.tgz` extracted into the source tree

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
* How CLI plugins work (see the [CLI Plugins RFC](https://github.com/hasura/ndc-hub/blob/cli-guidelines/rfcs/0002-cli-guidelines.md))

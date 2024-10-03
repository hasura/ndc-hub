# RFC: Docker-based CLI Plugin

There is a new requirement for wrapped CLI plugins provided by connectors that expose arbitrary commands defined by the connector that the user can invoke via the `ddn` CLI. The `ddn` CLI will, under the covers, select the correct "CLI plugin" for the connector and its version. So, for example, the user may issue the following `ddn` command:

```
ddn connector plugin --connector ./connector.yaml -- <args>
```

where `<args>` is a set of command line arguments passed by the `ddn` CLI to the correct plugin. The correct plugin to use is defined by the specified connector's connector definition. Currently, connectors can only specify binary-based CLI plugins. This RFC adds support for Docker-based CLI plugins to the connector definition.

This RFC builds on top of previous RFCs:
* [Connector Packaging and Hub Connector Definition](https://github.com/hasura/ndc-hub/blob/main/rfcs/0001-packaging.md)
* [Native Packaging Definition](https://github.com/hasura/ndc-hub/blob/main/rfcs/0004-native-packaging.md)
* [CLI Plugin Guidelines](https://github.com/hasura/ndc-hub/blob/11be49aa671924c2f5ddcd18a3ff3cbfe9a2e5c3/rfcs/0002-cli-guidelines.md)

## Changes to the Connector Definition
The `connector-metadata.yaml` changes so that the `cliPlugin` property can now take a new "Docker" object variant that allows the specification of a Docker image to use. For example:

```yaml
packagingDefinition:
  type: PrebuiltDockerImage
  dockerImage: ghcr.io/hasura/ndc-mysql:v0.0.1
supportedEnvironmentVariables:
- name: CONNECTION_URI
  description: The MySQL connection URI
commands:
  update:
    type: Dockerized
    dockerImage: ghcr.io/hasura/ndc-mysql-cli-plugin:v0.0.1
    commandArgs: [ "update" ]
cliPlugin: # New Docker type
  type: Docker
  dockerImage: ghcr.io/hasura/ndc-mysql-cli-plugin:v0.0.1
dockerComposeWatch:
  - path: ./
    target: /etc/connector
    action: sync+restart
```

When the CLI wants to invoke this Dockerized CLI plugin, it simply looks for the Docker image specified at `cliPlugin.dockerImage` and runs it with the command args that are passed after the `--` in the original `ddn` command invocation. The same environment variables and mounted volumne should be applied to the Docker container as are applied when running a `DockerizedCommand` (ie. the `supportedEnvironmentVariables` as defined in the `ConnectorManifest` and the [env vars specified in the CLI plugin RFC](https://github.com/hasura/ndc-hub/blob/11be49aa671924c2f5ddcd18a3ff3cbfe9a2e5c3/rfcs/0002-cli-guidelines.md#inputs-to-the-plugin-invocation)).

### `connector-metadata.yaml` types

```typescript
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  nativeToolchainDefinition?: NativeToolchainDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
  commands: Commands
  cliPlugin?: BinaryCliPluginDefinition | DockerCliPluginDefinition // NEW: DockerCliPluginDefinition variant
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

type NativeToolchainDefinition = {
  commands: NativeToolchainCommands
}

type NativeToolchainCommands = {
  start: string | DockerizedCommand | ShellScriptCommand
  update?: string | DockerizedCommand | ShellScriptCommand
  watch: string | DockerizedCommand | ShellScriptCommand
}

type EnvironmentVariableDefinition = {
  name: string
  description: string defaultValue?: string
}

type Commands = {
  update?: string | DockerizedCommand | ShellScriptCommand
  watch?: string | DockerizedCommand | ShellScriptCommand
}

type DockerizedCommand = {
  type: "Dockerized"
  dockerImage: string // eg "hasura/postgres-data-connector:1.0.0"
  commandArgs: string[]
}

type ShellScriptCommand = {
  type: "ShellScript"
  bash: string
  powershell: string
}

type BinaryCliPluginDefinition = {
  type?: "Binary" // NEW! Optional property for backwards compatibility, should be specified with "Binary" from now on
  name: string
  version: string
}

// NEW!
type DockerCliPluginDefinition = {
  type: "Docker"
  dockerImage: string // eg "hasura/postgres-data-connector:1.0.0"
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

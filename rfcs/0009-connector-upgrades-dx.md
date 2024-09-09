# RFC: connector upgrades DX

How does a user upgrade across connector versions? For example: from postgres v0.7.0 to v1.1.1 ?

## Solution

`ddn connector upgrade --connector <path-to-connector.yaml>` will
- upgrade the connector.yaml and compose.yaml files in place
- atomically replace the .hasura-connector folder
- call a new plugin command `upgradeConfiguration`

The plugin command is expected to perform any connector specific upgrades to the connector configuration.

## `upgradeConfiguration` command

Should upgrade the context directory in place. The user is expected to use a version control system to handle changes.

Note: connectors should capture the configuration version as a part of the configuration.

Do not rely on the following files: connector.yaml, compose.yaml. Connectors must modify only files that the connector itself owns.


## `connector-metadata.yaml` types

```diff
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  nativeToolchainDefinition?: NativeToolchainDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
  commands: Commands
  cliPlugin?: CliPluginDefinition
  dockerComposeWatch: DockerComposeWatch
}

type PackagingDefinition = PrebuiltDockerImagePackaging | ManagedDockerBuildPackaging

type PrebuiltDockerImagePackaging = {
  type: "PrebuiltDockerImage"
  dockerImage: string 
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
  description: string
  defaultValue?: string
}

type Commands = {
  update?: string | DockerizedCommand | ShellScriptCommand 
  watch?: string | DockerizedCommand | ShellScriptCommand
  printSchemaAndCapabilities?: string |  DockerizedCommand | ShellScriptCommand
+ upgradeConfiguration?: string |  DockerizedCommand | ShellScriptCommand
}


type DockerizedCommand = {
  type: "Dockerized"
  dockerImage: string 
  commandArgs: string[]
}

type ShellScriptCommand = {
  type: "ShellScript"
  bash: string
  powershell: string
}

type CliPluginDefinition = {
  name: string
  version: string
}

type DockerComposeWatch = DockerComposeWatchItem[]

type DockerComposeWatchItem = {
  path: string
  action: "rebuild" | "sync" | "sync+restart"
  target?: string
  ignore?: string[]
}
```

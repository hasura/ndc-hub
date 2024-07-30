# RFC: `printSchema` command for all connectors

Currently, the only way to get the connector schema for CLI is to call `/schema` and `/capabilities` endpoints on a running NDC connector instance. In case of local connector build, CLI builds and runs the respective NDC connector in a docker container and calls the `/schema` and `/capabilities` endpoints on it. This is required by the CLI to build the DataConnectorLink which can then be used by the engine to query the underlying data source. The new local CLI workflow is aimed at reducing the number of steps to get a working API. In this workflow users should be able to update the DataConnectorLink without any dependency.

This RFC proposes to add a new optional command `printSchema` to the connector plugin which prints out NDC schema of the underlying data source(the type [VersionedSchemaAndCapabilities](https://github.com/hasura/v3-engine-multitenant/blob/c6da0685556f8f31194a7f6620513ff5c7261497/v3-metadata-build-service/jsonschemas/metadata.jsonschema#L5636)) to STDOUT . CLI can then call the plugin with relevant command to get the schema and construct DataConnectorLink from the output. All connector plugins are expected to have this command going forward. DDN CLI will use this to generate NDC schema if this command is present. In case, the command is not implemented on plugin (in older versions of connectors), CLI will rely on running the connector in docker to get the schema.

## Changes to the packaging spec
There is a new command `print-schema` in the `commands` section: 

```shell
packagingDefinition:
  type: ManagedDockerBuild/PrebuiltDockerImage
supportedEnvironmentVariables: []
commands:
  update: hasura-ndc-postgres update
  ####
  printSchema: hasura-ndc-plugin print-schema
  ####
```
The runtime behavior is the same as the other commands like `update`. CLI will set all the env vars required by the connector plugin (like `HASURA_DDN_CONNECTOR_CONTEXT_PATH`, `CONNECTION_URI`(for postgres connector), etc,.) and call this command. 

The equivalent plugin invocation can be done from DDN CLI as below:
```shell
ddn connector plugin --connector ./connector.yaml -- print-schema

// Prints out schema to STDOUT
```

## `connector-metadata.yaml` types

```typescript
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
  // ** NEW **
  printSchema?: string |  DockerizedCommand | ShellScriptCommand
  // ** END **
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

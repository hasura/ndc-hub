# RFC: support mandatory env vars in packaging spec

> [!NOTE]
> This RFC builds upon the concepts defined in [add documentation page to packaging spec](./0007-packaging-documentation-page.md)

Some environment variables are mandatory for certain connectors. These connectors may make a distinction
between the environment variable not being set and being set as an empty string.

If a user provides an empty string as value for an environment variable during `ddn connector init -i`, the
CLI does not include these variables in the environment variable mapping or the .env file. As a result, for
required environment variables, users must manually add them as empty strings for the connector to function correctly.

The CLI does not currently have a way to distinguish between optional environment variables where an empty value
should be ignored and a required environment variable where an empty value must still be written as an empty string.

This RFC aims to improve the user experience by eliminating the need for this manual step.

## Solution

This RFC proposes introducing a new optional field called `required` to indicate that an environment variable is mandatory. This way even if its value is empty, the CLI will still add it to the environment variable mapping and to the
.env file as an empty string so that the user does not have to do it manually.


## `connector-metadata.yaml` types

```typescript
type ConnectorMetadataDefinition = {
  packagingDefinition: PackagingDefinition
  nativeToolchainDefinition?: NativeToolchainDefinition
  supportedEnvironmentVariables: EnvironmentVariableDefinition[]
  commands: Commands
  cliPlugin?: CliPluginDefinition
  dockerComposeWatch: DockerComposeWatch
  documentationPage: string
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
  start: string | DockerizedCommand | ShellScriptCommand // Compulsory
  update?: string | DockerizedCommand | ShellScriptCommand
  watch: string | DockerizedCommand | ShellScriptCommand // Compulsory
}

type EnvironmentVariableDefinition = {
  name: string
  description: string
  defaultValue?: string
  required?: boolean // NEW
}

type Commands = {
  update?: string | DockerizedCommand | ShellScriptCommand
  watch?: string | DockerizedCommand | ShellScriptCommand
}

// From https://github.com/hasura/ndc-hub/pull/129 (outside of this RFC)
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

// From: https://github.com/compose-spec/compose-spec/blob/1938efd103f8e0817ca90e5f15177ec0317bbaf8/schema/compose-spec.json#L455
type DockerComposeWatch = DockerComposeWatchItem[]

type DockerComposeWatchItem = {
  path: string
  action: "rebuild" | "sync" | "sync+restart"
  target?: string
  ignore?: string[]
}
```
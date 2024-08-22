# RFC: add documentation page to packaging spec

> [!NOTE]
> This RFC builds upon the concepts defined in [Native Connector Packaging](./0004-native-packaging.md)

Connectors that run in the native environment such as typescript have a different getting started workflow for local development compared to connectors that run in docker such as postgres.

Users expect the same steps that they tried with one connector to work with another.

The CLI wishes to link users to the docs page for a specific connector on running `ddn connector init` so that users have clear directions to next steps.

## Solution

Extend the packaging spec definition to include `documentationPage` which points to a connector-specific getting started page. The primary use for this is for the CLI to link to getting started instructions after the user adds the connector to their project.

This should NOT point to the homepage or repository page for the connector, unless that page also contains getting started instructions. This should NOT point to any other reference pages aside from getting started instructions.

The page may be hosted on the Hasura docs, on Hasura hub, or on the repository. Hasura connector authors are encouraged to point to Hasura docs as this serves as the primary channel through which users are otherwise onboarded to ddn.

**connector-metadata.yaml**

```yaml
documentationPage: "https://hasura.io/docs/3.0/getting-started/build/add-business-logic?db=TypeScript"
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
  documentationPage: string // NEW
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
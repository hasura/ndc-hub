# RFC: Support Mandatory Environment Variables in Packaging Spec

> **Note:**  
> This RFC builds upon the concepts defined in [add documentation page to packaging spec](./0007-packaging-documentation-page.md).

Some environment variables are mandatory for certain connectors. These connectors may make a distinction between the environment variable not being set and being set as an empty string.

Currently, when a user provides an empty string for an environment variable during `ddn connector init -i`, the CLI does not include these variables in the environment variable mapping or the `.env` file. Users must manually add these environment variables to both the `.env` file and the mapping to ensure proper functionality.

The CLI does not currently have a way to distinguish between optional environment variables (where an empty value should be ignored) and required environment variables (where an empty value must be written as an empty string).

This RFC aims to improve the user experience by eliminating the need for this manual step.

## Proposed Solution

We propose adding a new optional field, `required`, to specify that an environment variable is mandatory. If this field is set, the CLI will include the environment variable in the mapping and `.env` file even if the user provides an empty string.

The table below describes whether the tooling writes the environment variable to the .env files or not based on the `defaultValue` and `required` fields:

| defaultValue/required | true                                                                                                                                | false\|undefined                                                                       |
|---------------------------|-------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------|
| Non-empty string          | YES                                                                                                                                 | YES (discouraged: tooling always writes default values even for non-required env vars) |
| Empty string              | YES                                                                                                                                 | NO (discouraged: inconsistent behavior) (backward-compatible behavior)              |
| Undefined                 | YES (discouraged: should provide a default value for required flags) (write user-provided string or empty string by default)       | NO                                                                                     |

#### Recommendations for Connector Authors

- **If the environment variable should be written to `.env` files:**
  - Mark the environment variable as required.
  - Provide a default value, even if it is an empty string.

- **If the environment variable should not be written to `.env` files:**
  - Do not mark the environment variable as required.
  - Do not provide a default value.

**Note:** Discouraged combinations are still valid and has a well defined behaviour but not recommended due to potential inconsistencies or behavior that may not align with best practices.


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

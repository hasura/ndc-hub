# RFC: Add Binary CLI plugin packaging information in `connector-metadata.yaml`

Currently, there are two types of connector CLI plugins:

1. Docker based CLI plugins
2. Binary CLI plugins

Some connectors choose to publish binary CLI plugins and part of the process of publishing a binary CLI plugin is to open a PR in the [cli-plugins-index](https://github.com/hasura/cli-plugins-index) and merge it to make it available in the DDN ecosystem.

This approach leads to poor user experience for two reasons. First, it's easy
to overlook the step of publishing the native CLI plugin. Second, this oversight
causes runtime errors during connector initialization. When initializing, the
connector attempts to introspect the data source. However, if the CLI plugin
isn't published, this introspection process fails

So, this RFC proposes to include the binary CLI plugin information within the `connector-metadata.yaml`,
so that all the information required to run a connector is in one place.

## Solution

Extend the connector packaging spec definition to include the binary CLI manifest([example](https://github.com/hasura/cli-plugins-index/blob/master/plugins/ndc-go/v1.4.0/manifest.yaml)) information in the `connector-metadata.yaml`.

The following structure is proposed:


### `connector-metadata.yaml` types

```typescript
export type ConnectorMetadataDefinition = {
  version?: "v1";
  packagingDefinition: PackagingDefinition;
  nativeToolchainDefinition?: NativeToolchainDefinition;
  supportedEnvironmentVariables: EnvironmentVariableDefinition[];
  commands: Commands;
  cliPlugin?: CliPluginDefinition;
  dockerComposeWatch: DockerComposeWatch;
  documentationPage?: string;
};

export type CliPluginDefinition =
  | BinaryCliPluginDefinition
  | DockerCliPluginDefinition;

export type PackagingDefinition =
  | PrebuiltDockerImagePackaging
  | ManagedDockerBuildPackaging;

export type PrebuiltDockerImagePackaging = {
  type: "PrebuiltDockerImage";
  dockerImage: string;
};

export type ManagedDockerBuildPackaging = {
  type: "ManagedDockerBuild";
};

export type NativeToolchainCommands = {
  start: string | DockerizedCommand | ShellScriptCommand;
  update?: string | DockerizedCommand | ShellScriptCommand;
  watch: string | DockerizedCommand | ShellScriptCommand;
};

export type NativeToolchainDefinition = {
  commands: NativeToolchainCommands;
};

export type Command = string | DockerizedCommand | ShellScriptCommand;

export type Commands = {
  update?: Command;
  watch?: Command;
  printSchemaAndCapabilities?: Command;
  upgradeConfiguration?: Command;
};

export type EnvironmentVariableDefinition = {
  name: string;
  description: string;
  defaultValue?: string;
  required?: string;
};

export type DockerizedCommand = {
  type: "Dockerized";
  dockerImage: string;
  commandArgs: string[];
};

export type ShellScriptCommand = {
  type: "ShellScript";
  bash: string;
  powershell: string;
};

export type DockerCliPluginDefinition = {
  type: "Docker";
  dockerImage: string; // eg "hasura/postgres-data-connector:1.0.0"
};

export type BinaryCliPluginPlatform = {
  /**
   * The selector identifies the target platform for this configuration.
   * It follows the format: <os>-<architecture>
   *
   * Possible values:
   * - darwin-arm64: macOS on ARM64 architecture (e.g., M1 Macs)
   * - linux-arm64: Linux on ARM64 architecture
   * - darwin-amd64: macOS on x86-64 architecture
   * - windows-amd64: Windows on x86-64 architecture
   * - linux-amd64: Linux on x86-64 architecture
   */
  selector:
    | "darwin-arm64"
    | "linux-arm64"
    | "darwin-amd64"
    | "windows-amd64"
    | "linux-amd64";
  /**
   * The URI of the CLI plugin archive, it should be in tar format.
   This should be a URL from which the binary can be downloaded,
   * without any authentication.
   */
  uri: string;
  /**
   * The SHA256 hash of the binary file.
   This is used to verify the integrity of the downloaded binary.
   */
  sha256: string;
  /**
   * The name of the binary file. This is the name of the binary file that will be placed in the bin directory.
   */
  bin: string;
};

export type FilePath = {
  from: string;
  to: string;
};

export type BinaryCliPluginDefinition =
  | BinaryExternalCliPluginDefinition
  | BinaryInlineCliPluginDefinition;

export type BinaryInlineCliPluginDefinition = {
  type?: "BinaryInline";
  platforms?: BinaryCliPluginPlatform[];
};

// When this type is found, it will fetch the plugin definition from https://github.com/hasura/cli-plugins-index
export type BinaryExternalCliPluginDefinition = {
  type?: "Binary";
  name: string;
  version: string;
};

export type DockerComposeWatch = DockerComposeWatchItem[];

export type DockerComposeWatchItem = {
  path: string;
  action: "rebuild" | "sync" | "sync+restart";
  target?: string;
  ignore?: string[];
};

```


### Summary of changes

- The `uri` specified in the `BinaryCliPluginPlatform` should download a tar package, where all the files required to run the plugin are present. For example, you may need to provide some `.dll` files for the windows binary to work correctly, supporting header files for a plugin which is not statically linked.

- Add a new optional field `version` to the `ConnectorMetadataDefinition` type. This field will be used to version the connector metadata definition.
- Introduce a new type `BinaryInlineCliPluginDefinition` to specify the binary CLI plugin information in the `connector-metadata.yaml`, this type contains the platform information.

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
  watch?: string | DockerizedCommand | ShellScriptCommand;
  upgradeConfiguration?: string | DockerizedCommand | ShellScriptCommand;
  cliPluginEntrypoint?: string | DockerizedCommand | ShellScriptCommand;
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
   * The URI of the CLI plugin.
   * This CLI binary plugin should be a URL from where the binary can be downloaded,
   * without any authentication.
   */
  uri: string;
  /**
   * The SHA256 hash of the binary file. This is used to verify the integrity of the downloaded binary.
   */
  sha256: string;
  /**
   * The name of the binary file. The binary file downloaded from the `uri` will be saved with this name.
   */
  bin: string;
};

export type BinaryCliPluginDefinition =
  | BinaryExternalCliPluginDefinition
  | BinaryInlineCliPluginDefinition;

export type BinaryInlineCliPluginDefinition = {
  type: "BinaryInline";
  platforms: BinaryCliPluginPlatform[];
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

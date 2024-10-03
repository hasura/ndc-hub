export type ConnectorMetadataDefinition = {
  version?: 1;
  packagingDefinition: PackagingDefinition;
  nativeToolchainDefinition?: NativeToolchainDefinition;
  supportedEnvironmentVariables: EnvironmentVariableDefinition[];
  commands: Commands;
  cliPlugin?: BinaryCliPluginDefinition | DockerCliPluginDefinition;
  dockerComposeWatch: DockerComposeWatch;
  documentationPage?: string;
};

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

export type Commands = {
  update?: string | DockerizedCommand | ShellScriptCommand;
  watch?: string | DockerizedCommand | ShellScriptCommand;
  printSchemaAndCapabilities?: string | DockerizedCommand | ShellScriptCommand;
  upgradeConfiguration?: string | DockerizedCommand | ShellScriptCommand;
};

export type EnvironmentVariableDefinition = {
  name: string;
  description: string;
  defaultValue?: string;
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
   * The URI of the binary file. This should be a URL from which the binary can be downloaded,
   * without any authentication.
   */
  uri: string;
  /**
   * The SHA256 hash of the binary file. This is used to verify the integrity of the downloaded binary.
   */
  sha256: string;
  /**
   * The name of the binary file. This is the name of the binary file that will be placed in the bin directory.
   */
  bin: string;

  files: File[];
};

export type FilePath = {
  from: string;
  to: string;
};

export type BinaryCliPluginDefinition = {
  type: "Binary";
  name: string;
  platforms: BinaryCliPluginPlatform[];
};

export type DockerComposeWatch = DockerComposeWatchItem[];

export type DockerComposeWatchItem = {
  path: string;
  action: "rebuild" | "sync" | "sync+restart";
  target?: string;
  ignore?: string[];
};

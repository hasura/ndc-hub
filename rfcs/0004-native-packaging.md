# RFC: Native Connector Packaging

> [!NOTE]
> This RFC builds upon the concepts defined in the [Deployment Packaging RFC](./0001-packaging.md)

Currently, all connectors execute inside Docker when running on the end user's machine. This was done to ensure that any end user could run the connector without requiring tooling other than Docker to be installed. It also meant that connectors were built and run in a reproducible way that was guaranteed to work the same when running in Hasura Cloud.

However, certain categories of connectors may benefit from being able to run natively on the user's machine, outside of Docker. In particular, connectors that involve some sort of "build" process may be best executed natively, for example the [NodeJS Lambda Connector](https://github.com/hasura/ndc-nodejs-lambda) or the [Go Connector](https://github.com/hasura/ndc-sdk-go/tree/main/cmd/ndc-go-sdk). 

This is because these connectors usually involve a developer toolchain and external tooling to use them effectively. For example, for the NodeJS Lambda Connector, it is expected that the user will edit their functions code using an editor such as VSCode. Doing so requires them to have the NodeJS toolchain installed locally and all the npm packages referenced in their `package.config` `npm install`ed locally. The best end user experience for these sorts of connectors may involve using hot-reloading, file system watching and code generation.

Unfortunately, when the connector only runs in Docker, the connector may execute correctly but the local tooling may show errors because the local toolchain is misconfigured or not installed. This leads to a disconnect where the user is seeing errors in one place but not another, because two toolchains are in use: one Dockerized, the other running locally. Also, achieving features like hot-reloading and file system watching can be more difficult to achieve when the connector is running inside Docker. This is because file system watchers do not work well through Docker volume mounts, and Docker Compose Watch's file sync functionality is one way (from the host into the container), so generated files inside the container are not copied to the host.

For connectors with these requirements (typically connectors that use the existing `ManagedDockerBuild` packaging), it may be preferrable to run these connectors directly on the end user's machine using the same toolchain they will edit the connector's files (such as their `functions.ts`) with. This should lead to a more consistent and predictable user experience.

## Use Cases
* `ddn dev` 
  * As an end user, when I run `ddn dev` I want my connector to start immediately using its current configuration, so that I can use my GraphQL API
  * As an end user, when I run `ddn dev` and I don't have the correct toolchain installed on my machine, I want the error messages to explain what I don't have installed clearly, so that I can correct the problem myself easily
  * As an end user, during `ddn dev` I want my connector to update with my latest changes, so that as I make changes I can see them reflected in my GraphQL API
* `ddn build connector-manifest` 
  * As an end user, when I build my connector manifest, I want to see any errors that would prevent my connector from running, so that I may correct them

## Proposed Solution
The proposed solution is to extend the connector definition so that it can capture whether or not a `ManagedDockerBuild` connector supports being run using the native toolchain.

The following changes have been made to the `connector-metadata.yaml`:
* `supportsNativeToolchain` has been added to `ManagedDockerBuildPackaging`. This allows the connector to declare whether or not it supports being run with a native toolchain. If this is set to `true` (defaults to `false`), the connector should also define a `Commands.init` and a `Commands.watch`. When the `ddn` CLI runs the connector during `ddn dev` it will do so using the `watch` command, rather than running a Docker container.
* `init` has been added to `Commands`. `init` is run by the `ddn` CLI before it either builds the connector (`ddn build connector-manifest`) or runs it (`ddn watch`). The purpose of `init` is to check the user's environment to ensure the native toolchain is installed and is ready to use. This includes restoring any dependencies, if necessary.
* A command type of `ShellScriptCommand` has been added that can be used for all `Commands`. A `ShellScriptCommand` defines a command that is executed using a shell script. There are two shells required, Bash (Mac/Linux) and PowerShell (Windows). When the specified command is run using the shell, the current working directory is set to where the contents of the `.hasura-connector` directory from the connector definition is located on disk. This enables the command to run a bundled script file. All the environment variables as specified in the CLI guidelines RFC will be set, including the `HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH`, which specifies where the build context directory is.
  * Note that for script files to be runnable in PowerShell, it must be invoked like so: `powershell.exe -ExecutionPolicy Bypass -Command <command>`. Running scripts is disabled by default on Windows clients. (For more information see [this documentation](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies?view=powershell-5.1).)

Here's a worked example for the NodeJS Lambda Connector:

**connector-metadata.yaml**

```yaml
packagingDefinition:
  type: ManagedDockerBuild
  supportsNativeToolchain: true
supportedEnvironmentVariables: {}
commands:
  init:
    type: ShellScript
    bash: ./init.sh
    powershell: ./init.ps1
  watch: npm run watch
```

**.hasura-connector/init.sh**
```bash
#!/usr/bin/env bash
set -eu -o pipefail

if ! command -v node &> /dev/null
then
  echo "node could not be found. Please install NodeJS v20+."
  exit 1
fi

# Potentially also check the node version here too

cd $HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH

# Restore packages
if [ ! -d "node_modules" ]
then
  npm ci
fi
```

**.hasura-connector/init.ps1**
```powershell
$ErrorActionPreference = "Stop"

if (-not (Get-Command "node" -ErrorAction SilentlyContinue)) {
  Write-Host "node could not be found. Please install NodeJS v20+."
  exit 1
}

# Potentially also check the node version here too

Set-Location $env:HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH

# Restore packages
if (Test-Path "./node_modules" -eq $false) {
  & npm ci
}
```

### `connector-metadata.yaml` changes

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
  supportsNativeToolchain?: boolean // *** NEW! ***
}

type EnvironmentVariableDefinition = {
  name: string
  description: string
  defaultValue?: string
}

type Commands = {
  init?: string | DockerizedCommand | ShellScriptCommand // *** NEW! ***
  update?: string | DockerizedCommand | ShellScriptCommand // *** Supports ShellScriptCommand now ***
  watch?: string | DockerizedCommand | ShellScriptCommand // *** Supports ShellScriptCommand now ***
}

type DockerizedCommand = {
  type: "Dockerized"
  dockerImage: string // eg "hasura/postgres-data-connector:1.0.0"
  commandArgs: string[]
}

// *** NEW! ***
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

## Alternative Solutions
### Embedded Shell
The most unappealing part of the proposed solution is the need to write two shell scripts if using `ShellScriptCommand`, in order to support both Linux/Mac (Bash) and Windows (PowerShell). We could try embedding a shell interpreter in the `ddn` CLI and using that to execute a shell script instead. For example, the [`interp`](https://pkg.go.dev/mvdan.cc/sh/v3@v3.8.0/interp) package purports to support a basic POSIX shell (with some Bash support).

However, this is unproven (especially on Windows), and risks compatibility issues where written scripts fail to execute correctly.

If we supported this, we could modify `ShellScriptCommand` to only allow execution via this embedded shell:

```typescript
type ShellScriptCommand = {
  type: "ShellScript"
  goInterp: string
}
```

## Open Questions
* What happens if the user runs `ddn build connector-manifest`?
* Is the proposed solution workable for the Go functions connector?
* Are there other things we'd rather do with an "init" command in the future? For example, delegate some of `ddn quickstart` process to the individual connectors so they can better validate user inputs such as connection strings etc? If so, how does that affect this?
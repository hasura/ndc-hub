# RFC: Native Connector Packaging

> [!NOTE]
> This RFC builds upon the concepts defined in the [Native Packaging RFC](./0004-native-packaging.md)

Currently, Connectors are packaged and shipped as docker images and Connector CLI plugins can be binary, Docker based or native (via a shell script wrapper). To support running connectors and plugins in the [Native Workspace](https://github.com/hasura/ddn-workspace/tree/main/native), we need to support running connectors and plugins natively.

## Proposal

DDN CLI will now pass two environment variables to the native runtime commands:

 - `HASURA_DDN_NATIVE_CONNECTOR_DIR`: Path to the directory where the connector executable (or the needed files like `jar`, `.js`, etc,.) is present. 
 - `HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR`: Path to the directory where the connector CLI plugin executable (or the needed files like `jar`, `.js`, etc,.) is present. 

With these environment variables, connector authors can write scripts to run the plugin and the connector natively, without having to know where the executable is present and just use these env vars to template the directory of these artifacts.

Also, to make native runtime configuration to have feature parity with binary and docker based plugin commands, the following optional command keys are added to native toolchain commands:

 - `upgradeConfiguration` - Exact same [functionality](./0010-connector-upgrades-dx.md) as `updagradeConfiguration` in binary and docker based commands.
 - `cliPluginEntrypoint` - This is the entrypoint to invoke arbitrary CLI commands (like the `native-operation` command in ndc-postgres CLI plugin). CLI will invoke this entrypoint and append all user provided arguments to it with a space.

## Example

```diff
packagingDefinition:
  type: PrebuiltDockerImage
  dockerImage: ghcr.io/hasura/ndc-postgres:v1.2.0
+ nativeToolchainDefinition:
+   commands:
+     start:
+       type: ShellScript
+       bash: |
+         #!/usr/bin/env bash
+         set -eu -o pipefail        
+         HASURA_CONFIGURATION_DIRECTORY="$HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH" "$HASURA_DDN_NATIVE_CONNECTOR_DIR/ndc-postgres" serve
+       powershell: |
+         $ErrorActionPreference = "Stop"
+         $env:HASURA_CONFIGURATION_DIRECTORY="$env:HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH"; & "$env:HASURA_DDN_NATIVE_CONNECTOR_DIR\ndc-postgres.exe" serve
+     update:
+       type: ShellScript
+       bash: |
+         #!/usr/bin/env bash
+         set -eu -o pipefail
+         "$HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR/hasura-ndc-postgres" update
+       powershell: |
+         $ErrorActionPreference = "Stop"
+         & "$env:HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR\hasura-ndc-postgres.exe" update
+     upgradeConfiguration:
+       type: ShellScript
+       bash: |
+         #!/usr/bin/env bash
+         set -eu -o pipefail
+         "$HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR/hasura-ndc-postgres" upgrade
+       powershell: |
+         $ErrorActionPreference = "Stop"
+         & "$env:HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR\hasura-ndc-postgres.exe" upgrade
+     cliPluginEntrypoint:
+       type: ShellScript
+       bash: ./entrypoint.sh
+       powershell: ./entrypoint.ps1
supportedEnvironmentVariables:
  - name: CONNECTION_URI
    description: The PostgreSQL connection URI
    defaultValue: postgresql://read_only_user:readonlyuser@35.236.11.122:5432/v3-docs-sample-app
  - name: CLIENT_CERT
    description: The SSL client certificate (Optional)
    defaultValue: ""
  - name: CLIENT_KEY
    description: The SSL client key (Optional)
    defaultValue: ""
  - name: ROOT_CERT
    description: The SSL root certificate (Optional)
    defaultValue: ""
commands:
  update: hasura-ndc-postgres update
cliPlugin:
  name: ndc-postgres
  version: v1.2.0
dockerComposeWatch:
  - path: ./
    target: /etc/connector
    action: sync+restart
```

**.hasura-connector/entrypoint.sh**
```bash
#!/usr/bin/env bash
set -eu -o pipefail
"$HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR/hasura-ndc-postgres" "$@"
```
**.hasura-connector/entrypoint.ps1**
```powershell
$ErrorActionPreference = "Stop"
& "$env:HASURA_DDN_NATIVE_CONNECTOR_PLUGIN_DIR\hasura-ndc-postgres.exe" "$Args"
```
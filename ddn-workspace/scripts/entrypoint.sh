#!/usr/bin/env bash

set -eo pipefail

CODE_SERVER_PORT=${CODE_SERVER_PORT:-8123}

CODE_SERVER_CMD_OPTIONS=""
if [[ -n "$CODE_SERVER_PASSWORD" ]]; then
    echo "Starting code-server with password"
    export PASSWORD="$CODE_SERVER_PASSWORD"
fi
if [[  "$CODE_SERVER_NO_AUTH" == "true" ]]; then
    CODE_SERVER_CMD_OPTIONS="$CODE_SERVER_CMD_OPTIONS --auth none"
    echo "Starting code-server with no auth"
fi

if [[ -n "$DDN_WORKSPACE_ACCESS_TOKEN" ]]; then
    ddn auth login --access-token "$DDN_WORKSPACE_ACCESS_TOKEN" || echo "DDN CLI login failed. Continuing..."
fi

code-server --disable-telemetry --disable-update-check --app-name "Hasura DDN" --welcome-text "DDN Dev Native Workspace" --disable-getting-started-override $CODE_SERVER_CMD_OPTIONS --bind-addr 0.0.0.0:"$CODE_SERVER_PORT" "/workspace"

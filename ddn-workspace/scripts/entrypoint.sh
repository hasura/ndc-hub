#!/usr/bin/env bash

set -eo pipefail

# Used from https://github.com/cruizba/ubuntu-dind/blob/master/logger.sh
function INFO(){
    local function_name="${FUNCNAME[1]}"
    local msg="$1"
    timeAndDate=$(date)
    echo "[$timeAndDate] [INFO] [${0}] $msg"
}

export CODE_SERVER_PORT=${CODE_SERVER_PORT:-8124}
export VSCODE_PROXY_URI=${VSCODE_PROXY_URI:-"./proxy/{{port}}"}
CODE_SERVER_CMD_OPTIONS=""
if [[ -n "$CODE_SERVER_PROXY_DOMAIN" ]]; then
    CODE_SERVER_CMD_OPTIONS="$CODE_SERVER_CMD_OPTIONS --proxy-domain $CODE_SERVER_PROXY_DOMAIN"
fi
if [[ -n "$CODE_SERVER_TRUSTED_ORIGINS" ]]; then
    CODE_SERVER_CMD_OPTIONS="$CODE_SERVER_CMD_OPTIONS --trusted-origins $CODE_SERVER_TRUSTED_ORIGINS"
fi
if [[ -n "$CODE_SERVER_PASSWORD" ]]; then
    INFO "Starting code-server with password"
    export PASSWORD="$CODE_SERVER_PASSWORD"
fi
if [[  "$CODE_SERVER_NO_AUTH" == "true" ]]; then
    CODE_SERVER_CMD_OPTIONS="$CODE_SERVER_CMD_OPTIONS --auth none"
    INFO "Starting code-server with no auth"
fi
export CODE_SERVER_CMD_OPTIONS

# Add pipx to PATH
export PATH="$HOME/.local/bin:$PATH"

INFO "Starting supervisor"
supervisord -c /etc/supervisor/supervisord.conf -n &
SUPERVISOR_PID=$!

# Wait for docker to be ready
INFO "Waiting for docker to be running"
while ! docker info >/dev/null 2>&1; do sleep 1; done
INFO "Docker running successfully, loading images"

for file in "/images"/*.tar; do
  if [[ -f "$file" ]]; then    
    docker load -i "$file"
  fi
done

INFO "Services started successfully..."

if [[ -n "$DDN_WORKSPACE_ACCESS_TOKEN" ]]; then
    ddn auth login --access-token "$DDN_WORKSPACE_ACCESS_TOKEN" || echo "DDN CLI login failed. Continuing..."
fi

wait $SUPERVISOR_PID

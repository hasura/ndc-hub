#!/usr/bin/env bash

set -euo pipefail

# https://github.com/coder/code-server
curl -fsSL https://code-server.dev/install.sh | sh -s -- --version="$1"

# Install Hasura LSP
code-server --install-extension "HasuraHQ.hasura"
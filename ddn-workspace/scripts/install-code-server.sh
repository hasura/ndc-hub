#!/usr/bin/env bash

set -euo pipefail

# https://coder.com/docs/code-server/install#standalone-releases
VERSION="$1"
curl -fL https://github.com/coder/code-server/releases/download/v$VERSION/code-server-$VERSION-linux-amd64.tar.gz \
  | tar -C "$LIB_PATH" -xz
mv "$LIB_PATH/code-server-$VERSION-linux-amd64" "$LIB_PATH/code-server-$VERSION"
ln -s "$LIB_PATH/code-server-$VERSION/bin/code-server" "$BIN_PATH/code-server"

# Install Hasura LSP
code-server --install-extension "HasuraHQ.hasura"

# Modify code-server's lib VSCode's extension package.json as it results in a false positive Vulnerability report
# https://linear.app/hasura/issue/DX-880/investigate-bofa-security-vulnerabilities-related-to-workspace#comment-84349b99
extensions=($(ls -d "$LIB_PATH/code-server-$VERSION/lib/vscode/extensions"/*))
for extension in "${extensions[@]}"; do
    packageJSON="$extension/package.json"
    if [ -f "$packageJSON" ]; then
        echo "Processing $packageJSON..."
        yq --inplace --indent 4 '.version = "100.0.0"' "$packageJSON"
        echo "Updated $packageJSON"
    else
        echo "Warning: $packageJSON does not exist. Skipping."
    fi
done


# Modify code-server's lib VSCode's package.json as it results in a false positive Vulnerability report (This issue is that when coder/code-server is built, it builds VSCODE with the original tag, but calls it code-server - https://github.com/coder/code-server/blob/main/ci/build/build-vscode.sh)
# https://linear.app/hasura/issue/DX-878/review-list-of-vulnerabilities-for-control-plane-from-db
yq --inplace --indent 4 ".version = \"$VERSION\"" "$LIB_PATH/code-server-$VERSION/lib/vscode/package.json"

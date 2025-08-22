#!/bin/bash

# Script to update specific connector versions in the JSON file
# Usage: ./update-connector-version.sh <connector-name> <version> [versions-file]
# Example: ./update-connector-version.sh postgres v2.1.2

set -euo pipefail

CONNECTOR_NAME="$1"
NEW_VERSION="$2"
VERSIONS_FILE="${3:-connector-versions.json}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[UPDATE] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[UPDATE] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[UPDATE] ERROR: $1${NC}"
    exit 1
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    error "jq is required but not installed. Please install jq first."
fi

# Check if versions file exists
if [[ ! -f "$VERSIONS_FILE" ]]; then
    error "Versions file '$VERSIONS_FILE' not found."
fi

# Check if connector exists in the JSON
log "Checking if connector '$CONNECTOR_NAME' exists in $VERSIONS_FILE"
log "Available connectors: $(jq -r 'keys | join(", ")' "$VERSIONS_FILE")"

if ! jq -e --arg key "$CONNECTOR_NAME" 'has($key)' "$VERSIONS_FILE" > /dev/null 2>&1; then
    error "Connector '$CONNECTOR_NAME' not found in $VERSIONS_FILE"
fi

# Get current version
CURRENT_VERSION=$(jq -r --arg key "$CONNECTOR_NAME" '.[$key]' "$VERSIONS_FILE")

log "Updating $CONNECTOR_NAME from $CURRENT_VERSION to $NEW_VERSION"

# Update the JSON file with better error handling
if ! jq --arg key "$CONNECTOR_NAME" --arg val "$NEW_VERSION" '.[$key] = $val' "$VERSIONS_FILE" > "${VERSIONS_FILE}.tmp"; then
    error "Failed to update JSON with jq"
fi

# Check if the temp file was created and has content
if [[ ! -s "${VERSIONS_FILE}.tmp" ]]; then
    error "Temporary file is empty or wasn't created"
fi

# Verify the update worked in the temp file
TEMP_VERSION=$(jq -r --arg key "$CONNECTOR_NAME" '.[$key]' "${VERSIONS_FILE}.tmp")
if [[ "$TEMP_VERSION" != "$NEW_VERSION" ]]; then
    error "Version update failed - expected $NEW_VERSION but got $TEMP_VERSION"
fi

# Move the temp file to replace the original
if ! mv "${VERSIONS_FILE}.tmp" "$VERSIONS_FILE"; then
    error "Failed to replace original file"
fi

log "Successfully updated $CONNECTOR_NAME to $NEW_VERSION in $VERSIONS_FILE"

# Show current state
echo ""
log "Current connector versions:"
jq -r 'to_entries[] | "  \(.key): \(.value)"' "$VERSIONS_FILE"

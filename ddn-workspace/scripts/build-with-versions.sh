#!/bin/bash

# Build script that reads connector versions from JSON and passes them as build args
# Usage: ./build-with-versions.sh [connector-versions.json] [image-tag] [ddn-environment]

set -euo pipefail

VERSIONS_FILE="${1:-connector-versions.json}"
IMAGE_TAG="${2:-ddn-workspace:latest}"
DDN_ENVIRONMENT="${3:-staging}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[BUILD] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[BUILD] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[BUILD] ERROR: $1${NC}"
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

log "Reading connector versions from $VERSIONS_FILE"

# Build the build args string from JSON
BUILD_ARGS=""
while IFS='=' read -r connector_name version; do
    if [[ -n "$connector_name" && -n "$version" ]]; then
        # Convert connector name to uppercase and replace hyphens with underscores
        ARG_NAME=$(echo "$connector_name" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
        
        BUILD_ARGS="$BUILD_ARGS --build-arg ${ARG_NAME}_VERSION=$version"
        log "Setting ${ARG_NAME}_VERSION=$version"
    fi
done < <(jq -r 'to_entries[] | "\(.key)=\(.value)"' "$VERSIONS_FILE")

log "Building Docker image: $IMAGE_TAG"
log "DDN Environment: $DDN_ENVIRONMENT"
log "Build args: $BUILD_ARGS"

# Execute docker build with the generated build args and DDN environment
eval "docker build $BUILD_ARGS --build-arg DDN_ENVIRONMENT=$DDN_ENVIRONMENT -t $IMAGE_TAG ."

if [[ $? -eq 0 ]]; then
    log "Successfully built image: $IMAGE_TAG"
    
    # Show the versions that were used
    echo ""
    log "Connector versions used in build:"
    jq -r 'to_entries[] | "  \(.key): \(.value)"' "$VERSIONS_FILE"
else
    error "Docker build failed"
fi

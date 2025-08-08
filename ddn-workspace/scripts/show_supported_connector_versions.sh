#!/usr/bin/env bash
set -euo pipefail

# Script to list all directories in {LIB_PATH} with their namespace, name, and version
# and output as JSON in the format {"namespace/name": "version", ...}

TEMP_FILE=$(mktemp)
echo "{}" > "$TEMP_FILE"

# Find all namespace/name/version directories
find "$LIB_PATH" -mindepth 3 -maxdepth 3 -type d | while read -r dir; do
  # Extract the relative path from LIB_PATH
  rel_path=${dir#"$LIB_PATH/"}
  
  # Extract namespace, name, and version
  namespace=$(echo "$rel_path" | cut -d '/' -f 1)
  name=$(echo "$rel_path" | cut -d '/' -f 2)
  version=$(echo "$rel_path" | cut -d '/' -f 3)
  
  # Check if version follows semver pattern (either v1.2.3 or 1.2.3)
  if [[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    # Create the key in the format "namespace/name"
    key="$namespace/$name"
    
    # Update the JSON object
    cat "$TEMP_FILE" | jq --arg key "$key" --arg val "$version" '. + {($key): $val}' > "$TEMP_FILE.new"
    mv "$TEMP_FILE.new" "$TEMP_FILE"
  fi
done

# Format and output the JSON
jq '.' "$TEMP_FILE"

# Clean up
rm "$TEMP_FILE"
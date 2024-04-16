#!/usr/bin/env bash

set -e
set -u
set -o pipefail

# Update ndc-postgres related entries with a new version.

if [ "$#" != 2 ];
then
  echo "Usage: $0 <new-tag> <new-hash>"
  exit 1
fi

TAG="$1"
HASH="$2"

# Update each ndc-postgres variant
for variant in \
    'aurora' \
    'citus' \
    'cockroach' \
    'neon' \
    'postgres' \
    'postgres-alloydb' \
    'postgres-azure' \
    'postgres-cosmos' \
    'postgres-gcp' \
    'postgres-timescaledb' ; do

# add new version
jq \
    --arg tag "${TAG}" \
    --arg hash "${HASH}" \
    '.source_code.version += [{"tag": $tag, "hash": $hash, "is_verified": true}]' \
    registry/"${variant}"/metadata.json \
    > registry/"${variant}"/metadata.json2

mv registry/"${variant}"/metadata.json2 registry/"${variant}"/metadata.json

# set latest version
jq \
    --arg tag "${TAG}" \
    '.overview.latest_version |= $tag' \
    registry/"${variant}"/metadata.json \
    > registry/"${variant}"/metadata.json2

mv registry/"${variant}"/metadata.json2 registry/"${variant}"/metadata.json

done

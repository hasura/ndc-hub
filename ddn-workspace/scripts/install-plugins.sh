#!/usr/bin/env bash

set -euo pipefail

if [[ -n "$1" ]]; then
    IFS=',' read -r -a plugins <<< "$1"
    for plugin in "${plugins[@]}"; do
        if [[ $plugin == *:* ]]; then
            name="${plugin%:*}"
            version="${plugin#*:}"
            ddn plugin install "$name" --version "$version"
        else
            ddn plugin install "$plugin"
        fi
    done
fi

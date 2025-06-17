#!/bin/sh
# env-template-wrapper.sh - Wrapper for env-template that parses URLs first
# Usage: ./env-template-wrapper.sh [env-template args...]

set -e

# Parse database URLs if they exist and export decomposed variables
if [ -f "./parse-db-url.sh" ]; then
    eval $(./parse-db-url.sh)
elif [ -f "/app/parse-db-url.sh" ]; then
    eval $(/app/parse-db-url.sh)
else
    echo "Warning: parse-db-url.sh not found, proceeding without URL parsing" >&2
fi

# Run env-template with the provided arguments
if [ -f "./env-template" ]; then
    exec ./env-template "$@"
elif [ -f "/app/env-template" ]; then
    exec /app/env-template "$@"
else
    echo "Error: env-template not found" >&2
    exit 1
fi
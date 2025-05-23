#!/bin/bash

# Exit on any error
set -e

# Check if environment argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <environment>"
    echo "Environment must be one of: poc, demo, development"
    exit 1
fi

ENV=$1
if [[ ! "$ENV" =~ ^(poc|demo|development)$ ]]; then
    echo "Error: Environment must be one of: poc, demo, development"
    exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../sql"


# Set up data
echo "Setting up fake data..."
cat "$SQL_DIR/fake_projmgmt_saas.sql" | docker exec -i ${ENV}-postgres-primary-1 psql -U postgres -d source_db

echo "...complete!" 
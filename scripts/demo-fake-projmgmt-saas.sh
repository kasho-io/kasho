#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
SQL_DIR="$PROJECT_ROOT/sql"

ENV=${1:-development}

ENV_FILE="$PROJECT_ROOT/environments/$ENV/.env"
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: Environment file $ENV_FILE not found"
    exit 1
fi
source "$ENV_FILE"

echo "Setting up fake data..."
cat "$SQL_DIR/demo/fake_projmgmt_saas.sql" | docker exec -i \
    -e PGUSER=${PRIMARY_DATABASE_SU_USER} \
    -e PGPASSWORD=${PRIMARY_DATABASE_SU_PASSWORD} \
    -e PGDATABASE=${PRIMARY_DATABASE_DB} \
    ${ENV}-postgres-primary-1 psql

echo "...complete!" 
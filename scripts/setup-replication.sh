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

if [ ! -f "$SQL_DIR/setup/10-setup-ddl-replication-log.sql" ]; then
    echo "Error: $SQL_DIR/setup/10-setup-ddl-replication-log.sql not found."
    exit 1
fi

if [ ! -f "$SQL_DIR/setup/20-setup-replication-slot.sql" ]; then
    echo "Error: $SQL_DIR/setup/20-setup-replication-slot.sql not found."
    exit 1
fi

echo "Setting up DDL replication log..."
cat "$SQL_DIR/setup/10-setup-ddl-replication-log.sql" | docker exec -i \
    -e PGUSER=${PRIMARY_DATABASE_SU_USER} \
    -e PGPASSWORD=${PRIMARY_DATABASE_SU_PASSWORD} \
    -e PGDATABASE=${PRIMARY_DATABASE_DB} \
    ${ENV}-postgres-primary-1 psql

echo "Setting up replication slot..."
cat "$SQL_DIR/setup/20-setup-replication-slot.sql" | docker exec -i \
    -e PGUSER=${PRIMARY_DATABASE_SU_USER} \
    -e PGPASSWORD=${PRIMARY_DATABASE_SU_PASSWORD} \
    -e PGDATABASE=${PRIMARY_DATABASE_DB} \
    ${ENV}-postgres-primary-1 psql

echo "Replication setup complete!" 
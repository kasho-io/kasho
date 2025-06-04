#!/bin/bash

# Exit on any error
set -e

# Get the absolute path of the script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../../../sql"

# Check if SQL files exist
if [ ! -f "$SQL_DIR/10-setup-ddl-replication-log.sql" ]; then
    echo "Error: $SQL_DIR/10-setup-ddl-replication-log.sql not found."
    exit 1
fi

if [ ! -f "$SQL_DIR/20-setup-replication-slot.sql" ]; then
    echo "Error: $SQL_DIR/20-setup-replication-slot.sql not found."
    exit 1
fi

echo "Setting up DDL replication log..."
cat "$SQL_DIR/10-setup-ddl-replication-log.sql" | docker exec -i development-postgres-primary-1 psql -U postgres -d primary_db

echo "Setting up replication slot..."
cat "$SQL_DIR/20-setup-replication-slot.sql" | docker exec -i development-postgres-primary-1 psql -U postgres -d primary_db

echo "Replication setup complete!" 
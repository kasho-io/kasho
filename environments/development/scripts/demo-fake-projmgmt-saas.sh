#!/bin/bash

# Exit on any error
set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../sql"

# Set up data
echo "Setting up fake data..."
cat "$SQL_DIR/fake_projmgmt_saas.sql" | docker exec -i development-postgres-primary-1 psql -U postgres -d primary_db

echo "...complete!" 
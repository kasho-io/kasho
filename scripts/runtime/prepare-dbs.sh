#!/bin/bash
# prepare-primary-db.sh - Prepare primary database for Kasho replication
# Requires superuser permissions on both primary and replica databases

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Parse database URLs to get individual components
eval $("$SCRIPT_DIR/parse-db-url.sh")

# Build superuser connection URLs
PRIMARY_SU_URL="postgresql://${PRIMARY_DATABASE_SU_USER}:${PRIMARY_DATABASE_SU_PASSWORD}@${PRIMARY_DATABASE_HOST}:${PRIMARY_DATABASE_PORT}/${PRIMARY_DATABASE_DB}"
REPLICA_SU_URL="postgresql://${REPLICA_DATABASE_SU_USER}:${REPLICA_DATABASE_SU_PASSWORD}@${REPLICA_DATABASE_HOST}:${REPLICA_DATABASE_PORT}/${REPLICA_DATABASE_DB}"

echo "=== Kasho Primary Database Setup ==="
echo "Primary: ${PRIMARY_DATABASE_HOST}:${PRIMARY_DATABASE_PORT}/${PRIMARY_DATABASE_DB}"
echo "Replica: ${REPLICA_DATABASE_HOST}:${REPLICA_DATABASE_PORT}/${REPLICA_DATABASE_DB}"
echo ""

# Step 1: Verify prerequisites
echo "1. Verifying prerequisites..."
if ! psql "$PRIMARY_SU_URL" -f "$PROJECT_ROOT/sql/setup/verify-prerequisites.sql"; then
    echo "ERROR: Prerequisites check failed. Please ensure:"
    echo "  - wal_level = logical in postgresql.conf"
    echo "  - PostgreSQL has been restarted after configuration changes"
    exit 1
fi

# Step 2: Create Kasho user on primary (read-only + replication)
echo ""
echo "2. Creating Kasho user on primary database..."
if ! (eval $("$SCRIPT_DIR/parse-db-url.sh") && "$PROJECT_ROOT/bin/env-template" < "$PROJECT_ROOT/sql/setup/create-kasho-user-primary.sql.template") | psql "$PRIMARY_SU_URL"; then
    echo "ERROR: Failed to create Kasho user on primary"
    exit 1
fi

# Step 3: Create Kasho user on replica (read-write)
echo ""
echo "3. Creating Kasho user on replica database..."
if ! (eval $("$SCRIPT_DIR/parse-db-url.sh") && "$PROJECT_ROOT/bin/env-template" < "$PROJECT_ROOT/sql/setup/create-kasho-user-replica.sql.template") | psql "$REPLICA_SU_URL"; then
    echo "ERROR: Failed to create Kasho user on replica"
    exit 1
fi

# Step 4: Set up DDL logging (before replication!)
echo ""
echo "4. Setting up DDL logging..."
if ! psql "$PRIMARY_SU_URL" -f "$PROJECT_ROOT/sql/setup/setup-ddl-logging.sql"; then
    echo "ERROR: Failed to set up DDL logging"
    exit 1
fi

# Step 5: Set up replication publication
echo ""
echo "5. Setting up replication publication..."
if ! psql "$PRIMARY_SU_URL" -f "$PROJECT_ROOT/sql/setup/setup-replication.sql"; then
    echo "ERROR: Failed to set up replication publication"
    exit 1
fi

echo ""
echo "=== Primary database setup complete! ==="
echo ""
echo "Next steps:"
echo "1. Run ./scripts/bootstrap-kasho.sh to create the replication slot and bootstrap existing data"
echo "2. Ensure pg-change-stream is running (it will start in WAITING mode)"
echo "3. Ensure translicator is running to apply changes to the replica"
echo ""
echo "Note: The replication slot is now created during bootstrap to avoid WAL accumulation"
echo ""
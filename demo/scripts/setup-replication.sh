#!/bin/bash

# Exit on any error
set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../../sql"

echo "Setting up replication..."

# Step 1: Configure WAL settings on primary
echo "Configuring WAL settings on primary..."
cat "$SQL_DIR/setup_wal_primary.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db

# Step 2: Restart PostgreSQL on primary to ensure WAL settings take effect
echo "Restarting PostgreSQL on primary..."
docker restart demo-pg_primary-1

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until docker exec demo-pg_primary-1 pg_isready -U postgres; do
    echo "Waiting for PostgreSQL to be ready..."
    sleep 1
done

# Step 3: Set up replication on primary
echo "Setting up replication on primary..."
cat "$SQL_DIR/setup_replication_primary.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db

# Step 4: Set up replication on replica
echo "Setting up replication on replica..."
cat "$SQL_DIR/setup_replication_replica.sql" | docker exec -i demo-pg_replica-1 psql -U postgres -d replica_db

echo "Replication setup complete!" 
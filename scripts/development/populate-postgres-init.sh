#!/bin/sh
# populate-postgres-init.sh - Populate PostgreSQL init directories with SQL files

set -e

echo "Populating PostgreSQL init directories..."

# Create directories if they don't exist
mkdir -p /app/primary-init.d /app/replica-init.d

# Copy SQL files to init directories with proper numeric prefixes
# Primary database files
cp /app/sql/reset/reset-schema.sql /app/primary-init.d/00-reset-schema.sql
cp /app/sql/demo/fake_projmgmt_saas.sql /app/primary-init.d/10-demo-data.sql
cp /app/sql/setup/create-kasho-user-primary.sql.template /app/primary-init.d/20-create-kasho-user.sql.template
cp /app/sql/setup/setup-ddl-logging.sql /app/primary-init.d/30-setup-ddl-logging.sql
cp /app/sql/setup/setup-replication.sql /app/primary-init.d/40-setup-replication.sql

# Replica database files
cp /app/sql/reset/reset-schema.sql /app/replica-init.d/00-reset-schema.sql
cp /app/sql/setup/create-kasho-user-replica.sql.template /app/replica-init.d/10-create-kasho-user.sql.template

# Parse database URLs to get environment variables
eval $(/app/scripts/parse-db-url.sh)

# Process templates
/app/bin/env-template --dirs /app/primary-init.d,/app/replica-init.d

echo "PostgreSQL init directories populated successfully"
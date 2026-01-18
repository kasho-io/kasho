#!/bin/sh
# populate-mysql-init.sh - Populate MySQL init directories with SQL files

set -e

echo "Populating MySQL init directories..."

# Create directories if they don't exist
mkdir -p /app/primary-init.d /app/replica-init.d

# Clean up stale files from previous runs (preserves .gitignore)
find /app/primary-init.d -name "*.sql" -o -name "*.template" | xargs rm -f 2>/dev/null || true
find /app/replica-init.d -name "*.sql" -o -name "*.template" | xargs rm -f 2>/dev/null || true

# Copy SQL files to init directories with proper numeric prefixes
# Primary database files
cp /app/sql/mysql/reset/reset-schema.sql /app/primary-init.d/00-reset-schema.sql
cp /app/sql/mysql/demo/fake_projmgmt_saas.sql /app/primary-init.d/10-demo-data.sql
cp /app/sql/mysql/setup/create-kasho-user-primary.sql.template /app/primary-init.d/20-create-kasho-user.sql.template
# Note: MySQL does not need setup-ddl-logging.sql - DDL is captured via binlog automatically
cp /app/sql/mysql/setup/setup-replication.sql /app/primary-init.d/30-setup-replication.sql

# Replica database files
cp /app/sql/mysql/reset/reset-schema.sql /app/replica-init.d/00-reset-schema.sql
cp /app/sql/mysql/setup/create-kasho-user-replica.sql.template /app/replica-init.d/10-create-kasho-user.sql.template

# Parse database URLs to get environment variables
eval $(/app/scripts/parse-db-url.sh)

# Process templates
/app/bin/env-template --dirs /app/primary-init.d,/app/replica-init.d

echo "MySQL init directories populated successfully"

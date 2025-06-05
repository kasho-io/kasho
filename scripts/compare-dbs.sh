#!/bin/bash

ENV=${1:-development}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
ENV_FILE="$PROJECT_ROOT/environments/$ENV/.env"

if [ ! -f "$ENV_FILE" ]; then
    echo "❌ Environment file not found: $ENV_FILE"
    exit 1
fi
source "$ENV_FILE"

TEMP_DIR="$SCRIPT_DIR/../tmp"
mkdir -p "$TEMP_DIR"
PRIMARY_DUMP="$TEMP_DIR/primary_dump.sql"
REPLICA_DUMP="$TEMP_DIR/replica_dump.sql"

echo "📥 Dumping users table schema from primary database..."
docker exec -i \
  -e PGUSER=${PRIMARY_DATABASE_SU_USER} \
  -e PGPASSWORD=${PRIMARY_DATABASE_SU_PASSWORD} \
  -e PGDATABASE=${PRIMARY_DATABASE_DB} \
  ${ENV}-postgres-primary-1 pg_dump --no-owner --no-acl --schema-only -t users > "$PRIMARY_DUMP"

echo "📥 Dumping users table schema from replica database..."
docker exec -i \
  -e PGUSER=${REPLICA_DATABASE_SU_USER} \
  -e PGPASSWORD=${REPLICA_DATABASE_SU_PASSWORD} \
  -e PGDATABASE=${REPLICA_DATABASE_DB} \
  ${ENV}-postgres-replica-1 pg_dump --no-owner --no-acl --schema-only -t users > "$REPLICA_DUMP"

echo "🔍 Comparing users table schemas..."
if diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP" > /dev/null; then
    echo "✅ Schema comparison: users tables have identical schemas"
else
    echo "❌ Schema comparison: Differences found between users tables"
    echo "Showing differences:"
    diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP"
fi

echo "📥 Dumping users table data from primary database..."
docker exec -i \
  -e PGUSER=${PRIMARY_DATABASE_SU_USER} \
  -e PGPASSWORD=${PRIMARY_DATABASE_SU_PASSWORD} \
  -e PGDATABASE=${PRIMARY_DATABASE_DB} \
  ${ENV}-postgres-primary-1 pg_dump --no-owner --no-acl --data-only --inserts --column-inserts -t users | sort > "$PRIMARY_DUMP"

echo "📥 Dumping users table data from replica database..."
docker exec -i \
  -e PGUSER=${REPLICA_DATABASE_SU_USER} \
  -e PGPASSWORD=${REPLICA_DATABASE_SU_PASSWORD} \
  -e PGDATABASE=${REPLICA_DATABASE_DB} \
  ${ENV}-postgres-replica-1 pg_dump --no-owner --no-acl --data-only --inserts --column-inserts -t users | sort > "$REPLICA_DUMP"

echo "🔍 Comparing users table data..."
if diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP" > /dev/null; then
    echo "✅ Data comparison: users tables have identical data"
else
    echo "❌ Data comparison: Differences found between users tables"
    echo "Showing differences:"
    diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP"
fi

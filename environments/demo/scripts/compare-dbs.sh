#!/bin/bash
# Remove set -e to continue even if there are differences
# set -e

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

# Create temporary directories for dumps
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEMP_DIR="$SCRIPT_DIR/../tmp"
mkdir -p "$TEMP_DIR"
PRIMARY_DUMP="$TEMP_DIR/primary_dump.sql"
REPLICA_DUMP="$TEMP_DIR/replica_dump.sql"

echo "📥 Dumping users table schema from primary database..."
docker exec -i ${ENV}-postgres-primary-1 pg_dump -U postgres -d source_db --no-owner --no-acl --schema-only -t users > "$PRIMARY_DUMP"

echo "📥 Dumping users table schema from replica database..."
docker exec -i ${ENV}-postgres-replica-1 pg_dump -U postgres -d replica_db --no-owner --no-acl --schema-only -t users > "$REPLICA_DUMP"

echo "🔍 Comparing users table schemas..."
if diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP" > /dev/null; then
    echo "✅ Schema comparison: users tables have identical schemas"
else
    echo "❌ Schema comparison: Differences found between users tables"
    echo "Showing differences:"
    diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP"
fi

# Now compare data
echo "📥 Dumping users table data from primary database..."
docker exec -i ${ENV}-postgres-primary-1 pg_dump -U postgres -d source_db --no-owner --no-acl --data-only --inserts --column-inserts -t users | sort > "$PRIMARY_DUMP"

echo "📥 Dumping users table data from replica database..."
docker exec -i ${ENV}-postgres-replica-1 pg_dump -U postgres -d replica_db --no-owner --no-acl --data-only --inserts --column-inserts -t users | sort > "$REPLICA_DUMP"

echo "🔍 Comparing users table data..."
if diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP" > /dev/null; then
    echo "✅ Data comparison: users tables have identical data"
else
    echo "❌ Data comparison: Differences found between users tables"
    echo "Showing differences:"
    diff -u "$PRIMARY_DUMP" "$REPLICA_DUMP"
fi

# Cleanup
#rm -rf "$TEMP_DIR" 
#!/bin/bash
# bootstrap-kasho-mysql.sh - Bootstrap Kasho replication from existing MySQL data

set -euo pipefail

# Configuration from environment or defaults
PRIMARY_DATABASE_URL="${PRIMARY_DATABASE_URL:-mysql://kasho:kasho@mysql-primary:3306/primary_db}"
KV_URL="${KV_URL:-redis://redis:6379}"
CHANGE_STREAM_SERVICE_ADDR="${CHANGE_STREAM_SERVICE_ADDR:-mysql-change-stream:50051}"

echo "=== Kasho MySQL Bootstrap Process ==="
echo "Primary database: $PRIMARY_DATABASE_URL"
echo "Redis: $KV_URL"
echo "Change stream service: $CHANGE_STREAM_SERVICE_ADDR"
echo ""

# Check if grpcurl is available
if ! command -v grpcurl &> /dev/null; then
    echo "ERROR: grpcurl is required but not installed"
    echo "Please install grpcurl: https://github.com/fullstorydev/grpcurl"
    exit 1
fi

# Check if mysql client is available
if ! command -v mysql &> /dev/null; then
    echo "ERROR: mysql client is required but not installed"
    exit 1
fi

# Check if mysqldump is available
if ! command -v mysqldump &> /dev/null; then
    echo "ERROR: mysqldump is required but not installed"
    exit 1
fi

# Parse database URL to get connection parameters
eval $(/app/scripts/parse-db-url.sh)

MYSQL_HOST="${PRIMARY_DATABASE_HOST:-}"
MYSQL_PORT="${PRIMARY_DATABASE_PORT:-}"
MYSQL_USER="${PRIMARY_DATABASE_KASHO_USER:-}"
MYSQL_PASSWORD="${PRIMARY_DATABASE_KASHO_PASSWORD:-}"
MYSQL_DATABASE="${PRIMARY_DATABASE_DB:-}"

if [[ -z "$MYSQL_HOST" ]]; then
    echo "ERROR: Failed to parse PRIMARY_DATABASE_URL"
    echo "URL: $PRIMARY_DATABASE_URL"
    exit 1
fi

# Check if mysql-change-stream is running and in WAITING state
echo "Checking mysql-change-stream status..."
STATUS=$(grpcurl -import-path /app/proto -proto change_stream.proto -plaintext "$CHANGE_STREAM_SERVICE_ADDR" change_stream.ChangeStream/GetStatus 2>&1)
if [[ $? -ne 0 ]]; then
    echo "ERROR: Cannot connect to mysql-change-stream at $CHANGE_STREAM_SERVICE_ADDR"
    echo "Please ensure mysql-change-stream is running"
    exit 1
fi

CURRENT_STATE=$(echo "$STATUS" | grep -o '"state": "[^"]*"' | cut -d'"' -f4)
if [[ "$CURRENT_STATE" != "WAITING" ]]; then
    echo "ERROR: mysql-change-stream is in $CURRENT_STATE state, expected WAITING"
    echo "Bootstrap can only be started from WAITING state"
    exit 1
fi

echo "mysql-change-stream is in WAITING state, ready for bootstrap"
echo ""

# Step 1: Get current binlog position
echo "1. Getting binlog position..."

# MySQL connection options - skip SSL for development (self-signed certs)
MYSQL_OPTS="--skip-ssl"

# Test basic connectivity
if ! mysql $MYSQL_OPTS -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; then
    echo "ERROR: Cannot connect to MySQL at $MYSQL_HOST:$MYSQL_PORT"
    mysql $MYSQL_OPTS -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "SELECT 1" 2>&1
    exit 1
fi

# Get binlog position (redirect stderr to avoid deprecation warnings polluting output)
BINLOG_OUTPUT=$(mysql $MYSQL_OPTS -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -N -e "SHOW MASTER STATUS" 2>/dev/null)

if [[ $? -ne 0 ]]; then
    echo "ERROR: Failed to get binlog position"
    mysql $MYSQL_OPTS -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -N -e "SHOW MASTER STATUS" 2>&1
    echo ""
    echo "The MySQL user may need additional privileges:"
    echo "  GRANT RELOAD, REPLICATION CLIENT ON *.* TO 'kasho'@'%';"
    exit 1
fi

BINLOG_INFO=$(echo "$BINLOG_OUTPUT" | head -1)

if [[ -z "$BINLOG_INFO" ]]; then
    echo "ERROR: Failed to get binlog position"
    echo "Make sure binary logging is enabled (--log-bin)"
    exit 1
fi

BINLOG_FILE=$(echo "$BINLOG_INFO" | awk '{print $1}')
BINLOG_POS=$(echo "$BINLOG_INFO" | awk '{print $2}')
START_POSITION="${BINLOG_FILE}:${BINLOG_POS}"

echo "   Binlog position: $START_POSITION"

# Step 2: Signal mysql-change-stream to start accumulating
echo ""
echo "2. Starting change accumulation..."
RESPONSE=$(grpcurl -import-path /app/proto -proto change_stream.proto -plaintext \
  -d "{\"start_position\": \"$START_POSITION\"}" \
  "$CHANGE_STREAM_SERVICE_ADDR" change_stream.ChangeStream/StartBootstrap 2>&1)

if [[ $? -ne 0 ]]; then
    echo "ERROR: Failed to start bootstrap"
    echo "$RESPONSE"
    exit 1
fi

echo "   Change stream is now accumulating changes"

# Step 3: Take database dump
echo ""
echo "3. Dumping database..."
DUMP_FILE="/tmp/kasho_mysql_bootstrap_$(date +%Y%m%d_%H%M%S).sql"

# mysqldump with options for consistent backup
# --single-transaction: Use a consistent snapshot for InnoDB tables
# --routines: Include stored procedures and functions
# --triggers: Include triggers
# --no-tablespaces: Don't include tablespace info (not needed for replica)
# --skip-lock-tables: Don't lock tables (--single-transaction handles consistency)
# --complete-insert: Include column names in INSERT statements (required for translicator)
# Note: --set-gtid-purged is MySQL-specific and not supported by MariaDB's mysqldump

set +e  # Temporarily disable exit on error
mysqldump \
    $MYSQL_OPTS \
    -h "$MYSQL_HOST" \
    -P "$MYSQL_PORT" \
    -u "$MYSQL_USER" \
    -p"$MYSQL_PASSWORD" \
    --single-transaction \
    --routines \
    --triggers \
    --no-tablespaces \
    --skip-lock-tables \
    --complete-insert \
    "$MYSQL_DATABASE" > "$DUMP_FILE" 2>"${DUMP_FILE}.err"

DUMP_EXIT_CODE=$?
set -e  # Re-enable exit on error

if [[ $DUMP_EXIT_CODE -ne 0 ]]; then
    echo "ERROR: Database dump failed (exit code: $DUMP_EXIT_CODE)"
    cat "${DUMP_FILE}.err" 2>/dev/null || true
    rm -f "${DUMP_FILE}.err"
    exit 1
fi
rm -f "${DUMP_FILE}.err"

DUMP_SIZE=$(du -h "$DUMP_FILE" | cut -f1)
echo "   Dump complete: $DUMP_FILE ($DUMP_SIZE)"

# Step 4: Process dump with mysql-bootstrap-sync
echo ""
echo "4. Processing dump with mysql-bootstrap-sync..."
/app/bin/mysql-bootstrap-sync \
  --dump-file="$DUMP_FILE" \
  --kv-url="$KV_URL" &

BOOTSTRAP_PID=$!

# Optional: Wait for completion
if [[ "${WAIT_FOR_BOOTSTRAP:-false}" == "true" ]]; then
    echo "   Waiting for bootstrap to complete..."
    wait $BOOTSTRAP_PID

    echo ""
    echo "5. Transitioning to streaming mode..."
    grpcurl -import-path /app/proto -proto change_stream.proto -plaintext "$CHANGE_STREAM_SERVICE_ADDR" change_stream.ChangeStream/CompleteBootstrap

    echo ""
    echo "=== Bootstrap complete! ==="
    echo "The system is now in streaming mode."
else
    echo "   Bootstrap sync running in background (PID: $BOOTSTRAP_PID)"
    echo ""
    echo "When complete, transition to streaming mode with:"
    echo "  grpcurl -import-path /app/proto -proto change_stream.proto -plaintext $CHANGE_STREAM_SERVICE_ADDR change_stream.ChangeStream/CompleteBootstrap"

    echo ""
    echo "=== Bootstrap process initiated ==="
    echo "Monitor progress with:"
    echo "  grpcurl -import-path /app/proto -proto change_stream.proto -plaintext $CHANGE_STREAM_SERVICE_ADDR change_stream.ChangeStream/GetStatus"
fi

echo ""

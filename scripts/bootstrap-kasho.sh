#!/bin/bash
# bootstrap-kasho.sh - Bootstrap Kasho replication from existing data

set -euo pipefail

# Configuration from environment or defaults
PRIMARY_DATABASE_URL="${PRIMARY_DATABASE_URL:-postgresql://kasho:kasho@postgres-primary:5432/primary_db?sslmode=disable}"
KV_URL="${KV_URL:-redis://redis:6379}"
CHANGE_STREAM_SERVICE="${CHANGE_STREAM_SERVICE:-pg-change-stream:8080}"
REPLICATION_SLOT_NAME="${REPLICATION_SLOT_NAME:-kasho_slot}"

echo "=== Kasho Bootstrap Process ==="
echo "Primary database: $PRIMARY_DATABASE_URL"
echo "Redis: $KV_URL"
echo "Change stream service: $CHANGE_STREAM_SERVICE"
echo ""

# Check if grpcurl is available
if ! command -v grpcurl &> /dev/null; then
    echo "ERROR: grpcurl is required but not installed"
    echo "Please install grpcurl: https://github.com/fullstorydev/grpcurl"
    exit 1
fi

# Check if pg-change-stream is running and in WAITING state
echo "Checking pg-change-stream status..."
STATUS=$(grpcurl -import-path /app/proto -proto change_stream.proto -plaintext "$CHANGE_STREAM_SERVICE" change_stream.ChangeStream/GetStatus 2>&1)
if [[ $? -ne 0 ]]; then
    echo "ERROR: Cannot connect to pg-change-stream at $CHANGE_STREAM_SERVICE"
    echo "Please ensure pg-change-stream is running"
    exit 1
fi

CURRENT_STATE=$(echo "$STATUS" | grep -o '"state": "[^"]*"' | cut -d'"' -f4)
if [[ "$CURRENT_STATE" != "WAITING" ]]; then
    echo "ERROR: pg-change-stream is in $CURRENT_STATE state, expected WAITING"
    echo "Bootstrap can only be started from WAITING state"
    exit 1
fi

echo "pg-change-stream is in WAITING state, ready for bootstrap"
echo ""

# Step 1: Create or verify the permanent replication slot
echo "1. Setting up replication slot..."

# Check if the slot already exists
EXISTING_SLOT=$(psql "$PRIMARY_DATABASE_URL" -t -A -c "
  SELECT slot_name || '|' || confirmed_flush_lsn 
  FROM pg_replication_slots 
  WHERE slot_name = '$REPLICATION_SLOT_NAME';
")

if [[ -n "$EXISTING_SLOT" ]]; then
    echo "Replication slot '$REPLICATION_SLOT_NAME' already exists"
    IFS='|' read -r SLOT_NAME START_LSN <<< "$EXISTING_SLOT"
    echo "Using existing slot with LSN: $START_LSN"
else
    echo "Creating new replication slot '$REPLICATION_SLOT_NAME'..."
    SLOT_INFO=$(psql "$PRIMARY_DATABASE_URL" -t -A -c "
      SELECT slot_name || '|' || lsn FROM pg_create_logical_replication_slot('$REPLICATION_SLOT_NAME', 'pgoutput');
    ")
    
    if [[ $? -ne 0 ]]; then
        echo "ERROR: Failed to create replication slot"
        exit 1
    fi
    
    IFS='|' read -r SLOT_NAME START_LSN <<< "$SLOT_INFO"
    echo "Created permanent slot: $SLOT_NAME"
    echo "Starting LSN: $START_LSN"
fi

# Step 2: Signal pg-change-stream to start accumulating
echo ""
echo "2. Starting change accumulation..."
RESPONSE=$(grpcurl -import-path /app/proto -proto change_stream.proto -plaintext \
  -d "{\"start_lsn\": \"$START_LSN\"}" \
  "$CHANGE_STREAM_SERVICE" change_stream.ChangeStream/StartBootstrap 2>&1)

if [[ $? -ne 0 ]]; then
    echo "ERROR: Failed to start bootstrap"
    echo "$RESPONSE"
    exit 1
fi

echo "Bootstrap started, pg-change-stream is now accumulating changes"

# Step 3: Take database dump
echo ""
echo "3. Dumping database (this may take a while)..."
DUMP_FILE="/tmp/kasho_bootstrap_$(date +%Y%m%d_%H%M%S).sql"
# pg_dump creates its own consistent snapshot internally
# Exclude objects that shouldn't be replicated:
# --no-publications: Exclude publication definitions (for logical replication)
# --no-subscriptions: Exclude subscription definitions (for logical replication)
# Note: Event triggers and regular triggers are included as they may be needed on replicas
if ! pg_dump "$PRIMARY_DATABASE_URL" \
  --no-owner \
  --no-privileges \
  --no-publications \
  --no-subscriptions \
  -f "$DUMP_FILE"; then
    echo "ERROR: Database dump failed"
    exit 1
fi

DUMP_SIZE=$(du -h "$DUMP_FILE" | cut -f1)
echo "Database dump complete: $DUMP_FILE ($DUMP_SIZE)"

# Step 4: No cleanup needed - we're using the permanent slot
echo ""
echo "4. Bootstrap data prepared"

# Step 5: Start pg-bootstrap-sync
echo ""
echo "5. Processing dump with pg-bootstrap-sync..."
/app/bin/pg-bootstrap-sync \
  --dump-file="$DUMP_FILE" \
  --kv-url="$KV_URL" &

BOOTSTRAP_PID=$!
echo "pg-bootstrap-sync running (PID: $BOOTSTRAP_PID)"

# Optional: Wait for completion
if [[ "${WAIT_FOR_BOOTSTRAP:-false}" == "true" ]]; then
    echo "Waiting for bootstrap to complete..."
    wait $BOOTSTRAP_PID
    
    echo ""
    echo "6. Transitioning to streaming mode..."
    grpcurl -import-path /app/proto -proto change_stream.proto -plaintext "$CHANGE_STREAM_SERVICE" change_stream.ChangeStream/CompleteBootstrap
    
    echo ""
    echo "=== Bootstrap complete! ==="
    echo ""
    echo "The system is now in streaming mode."
else
    echo ""
    echo "Bootstrap sync is running in background."
    echo "When complete, transition to streaming mode with:"
    echo "  grpcurl -import-path /app/proto -proto change_stream.proto -plaintext $CHANGE_STREAM_SERVICE change_stream.ChangeStream/CompleteBootstrap"
    
    echo ""
    echo "=== Bootstrap process initiated ==="
    echo ""
    echo "Monitor progress with:"
    echo "  grpcurl -import-path /app/proto -proto change_stream.proto -plaintext $CHANGE_STREAM_SERVICE change_stream.ChangeStream/GetStatus"
fi

echo ""
# Bootstrap and Stream Coordination Design

## Overview

This document describes the design for coordinating Kasho's initial bootstrap process with continuous streaming replication. The goal is to ensure no data loss during the transition from database snapshot to live replication.

## SQL Script Organization

### Current State

SQL scripts are currently scattered across multiple directories:
- `sql/setup/` - Core setup scripts with numeric prefixes (WAL level, DDL logging, replication slots)
- `sql/demo/` - Demo data for development
- `environments/development/primary-init.d/` - Primary database initialization with numeric prefixes
- `environments/development/replica-init.d/` - Replica database initialization with numeric prefixes
- Templates in various locations with environment variable substitution

### Proposed Consolidation

All SQL scripts should be consolidated in the `sql/` directory without numeric prefixes (except in init.d directories where execution order matters):

```
sql/
├── setup/                          # Core Kasho setup scripts (no numeric prefixes)
│   ├── verify-prerequisites.sql
│   ├── create-kasho-user.sql.template
│   ├── setup-ddl-logging.sql      # Must run before replication setup
│   ├── setup-replication.sql
│   └── setup-wal-level.sql
├── bootstrap/                      # Bootstrap-specific scripts
│   ├── create-snapshot.sql
│   ├── cleanup-snapshot.sql
│   └── README.md                  # DBA instructions
├── demo/                          # Demo/development data
│   └── fake-projmgmt-saas.sql
└── reset/                         # Database reset scripts
    └── reset-schema.sql
```

### Environment Integration with Proper Ordering

Development and demo environments will use symlinks with numeric prefixes to ensure proper execution order:

```
environments/development/
├── primary-init.d/
│   ├── 00-reset-schema.sql -> ../../../sql/reset/reset-schema.sql
│   ├── 10-demo-data.sql -> ../../../sql/demo/fake-projmgmt-saas.sql
│   ├── 20-create-kasho-user.sql -> ../../../sql/setup/create-kasho-user.sql
│   ├── 30-setup-ddl-logging.sql -> ../../../sql/setup/setup-ddl-logging.sql
│   └── 40-setup-replication.sql -> ../../../sql/setup/setup-replication.sql
└── replica-init.d/
    ├── 00-reset-schema.sql -> ../../../sql/reset/reset-schema.sql
    └── 10-create-kasho-user.sql -> ../../../sql/setup/create-kasho-user.sql
```

**Critical ordering requirements:**
1. WAL level must be configured in postgresql.conf (development: docker-compose command args, production: DBA configuration)
2. Demo data loaded before Kasho setup (mirrors production where data exists first)
3. Kasho user creation before any permission-dependent operations
4. DDL logging setup must happen before replication setup

This approach provides:
- Single source of truth for each script in `sql/` directory
- Clear execution ordering only where needed (init.d directories)
- Easy updates (change once, affects all environments)
- Simple deployment scripts can reference the canonical `sql/` directory

## pg-change-stream State Machine

### State Definitions

pg-change-stream will implement three operational states:

```
┌─────────────┐
│   WAITING   │ ← Waiting for bootstrap signal
└──────┬──────┘
       │ receive start_lsn
┌──────▼──────┐
│ ACCUMULATING│ ← Capturing changes, not streaming
└──────┬──────┘
       │ receive transition signal
┌──────▼──────┐
│  STREAMING  │ ← Normal operation
└─────────────┘
```

### State Behaviors

#### WAITING State
- Default startup state when no replication slot exists or database is unavailable
- Periodically attempts to connect and check for replication slot
- Remains in WAITING if: connection fails, user doesn't exist, or slot doesn't exist
- Does not start replication streaming
- Waits for bootstrap coordinator to provide starting LSN
- Exposes gRPC endpoint for bootstrap coordination

This allows pg-change-stream to start before database setup is complete

#### ACCUMULATING State
- Connects to PostgreSQL and starts replication from provided LSN
- Stores all changes in Redis
- Does NOT stream changes to clients (gRPC calls block/wait)
- Tracks metrics: changes accumulated, current LSN, lag

#### STREAMING State
- Normal operational mode
- Streams changes to connected clients
- Can resume from existing replication slot on restart

### State Persistence

State information is stored in Redis to survive pg-change-stream restarts:

```json
{
  "kasho:change-stream:state": {
    "current": "ACCUMULATING", 
    "start_lsn": "0/1234567"
  }
}
```

**Why Redis?**
- We're already using Redis to store changes
- Simple key-value storage for minimal state
- Updated only on state transitions (rare)
- If Redis fails, we restart bootstrap process (acceptable)

**Recovery on pg-change-stream restart:**
1. Check Redis for saved state
2. If ACCUMULATING: Resume from PostgreSQL replication slot's `restart_lsn` (≥ start_lsn)
3. If STREAMING: Resume normal operation from slot position  
4. If no state found: Start in WAITING mode

**LSN Relationship:**
- `start_lsn` (Redis): Snapshot boundary - don't process changes before this
- `restart_lsn` (PostgreSQL slot): Current position - resume replication from here
- Always: `restart_lsn` ≥ `start_lsn`

The PostgreSQL replication slot handles LSN position persistence. Redis just remembers which mode we're in and the original snapshot boundary.

### gRPC API

#### Service Definition

```protobuf
service ChangeStreamService {
  // Existing streaming method
  rpc StreamChanges(StreamRequest) returns (stream Change);
  
  // Bootstrap coordination methods
  rpc StartBootstrap(StartBootstrapRequest) returns (BootstrapResponse);
  rpc CompleteBootstrap(CompleteBootstrapRequest) returns (BootstrapResponse);
  rpc GetStatus(GetStatusRequest) returns (StatusResponse);
}

message StartBootstrapRequest {
  string start_lsn = 1;
  string snapshot_name = 2;
}

message CompleteBootstrapRequest {}

message GetStatusRequest {}

message BootstrapResponse {
  string status = 1;
  string previous_state = 2;
  string current_state = 3;
  int64 accumulated_changes = 4;
  bool ready_to_stream = 5;
}

message StatusResponse {
  string state = 1;
  string start_lsn = 2;
  string current_lsn = 3;
  int64 accumulated_changes = 4;
  int32 connected_clients = 5;
  int64 uptime_seconds = 6;
}
```

## Bootstrap Orchestration Process

### Automated Bootstrap Script

For ease of use, the bootstrap process can be automated with a shell script:

```bash
#!/bin/bash
# scripts/bootstrap-kasho.sh - Bootstrap script for Kasho replication

set -euo pipefail

# Configuration from environment or defaults
PRIMARY_DATABASE_URL="${PRIMARY_DATABASE_URL:-postgresql://kasho:kasho@postgres-primary:5432/primary_db?sslmode=disable}"
KV_URL="${KV_URL:-redis://redis:6379}"
CHANGE_STREAM_SERVICE="${CHANGE_STREAM_SERVICE:-pg-change-stream:8080}"
REPLICATION_SLOT_NAME="${REPLICATION_SLOT_NAME:-kasho_slot}"

echo "Starting Kasho bootstrap process..."

# Step 1: Create or verify the permanent replication slot
echo "Setting up replication slot..."

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
    
    IFS='|' read -r SLOT_NAME START_LSN <<< "$SLOT_INFO"
    echo "Created permanent slot: $SLOT_NAME"
    echo "Starting LSN: $START_LSN"
fi

# Step 2: Signal pg-change-stream to start accumulating
echo "Starting change accumulation..."
grpcurl -plaintext \
  -d "{\"start_lsn\": \"$START_LSN\"}" \
  "$CHANGE_STREAM_SERVICE" change_stream.ChangeStream/StartBootstrap

# Step 3: Take database dump
echo "Dumping database (this may take a while)..."
# Note: pg_dump creates its own consistent snapshot internally
pg_dump "$PRIMARY_DATABASE_URL" \
  --no-owner \
  --no-privileges \
  -f /tmp/bootstrap_dump.sql

# Step 4: No cleanup needed - we're using the permanent slot

# Step 5: Start pg-bootstrap-sync
echo "Processing dump into change events..."
pg-bootstrap-sync \
  --dump-file=/tmp/bootstrap_dump.sql \
  --redis-url="$KV_URL" &

BOOTSTRAP_PID=$!

# Step 6: Monitor progress
echo "Bootstrap sync running (PID: $BOOTSTRAP_PID)"
echo "Monitor progress with: docker logs <container>"

# Wait for completion or allow running in background
if [[ "${WAIT_FOR_COMPLETION:-false}" == "true" ]]; then
  wait $BOOTSTRAP_PID
  echo "Bootstrap complete!"
  
  # Optional: Signal streaming mode
  # grpcurl -plaintext "$CHANGE_STREAM_HOST" change_stream.ChangeStream/CompleteBootstrap
fi
```

To use this script:

```bash
# Copy script into container (or build into image)
docker cp scripts/bootstrap-kasho.sh kasho:/usr/local/bin/

# Run bootstrap process  
docker exec -it kasho bootstrap-kasho.sh

# Or run with custom configuration
docker exec -it kasho \
  -e PRIMARY_DATABASE_URL="postgresql://kasho:kasho@primary:5432/mydb?sslmode=disable" \
  -e KV_URL="redis://redis:6379" \
  -e CHANGE_STREAM_SERVICE="pg-change-stream:8080" \
  bootstrap-kasho.sh

# Note: Replace 'kasho' with your actual container name if different
```

### DBA Setup Script

For initial database setup, DBAs can use this script that runs the SQL files in order:

```bash
#!/bin/bash
# scripts/setup-kasho-db.sh - Run Kasho setup SQL scripts (requires SUPERUSER)

set -euo pipefail

# Parse database URLs to get individual components
eval $(./scripts/parse-db-url.sh)

# Build superuser connection URLs
PRIMARY_SU_URL="postgresql://${PRIMARY_DATABASE_SU_USER}:${PRIMARY_DATABASE_SU_PASSWORD}@${PRIMARY_DATABASE_HOST}:${PRIMARY_DATABASE_PORT}/${PRIMARY_DATABASE_DB}"
REPLICA_SU_URL="postgresql://${REPLICA_DATABASE_SU_USER}:${REPLICA_DATABASE_SU_PASSWORD}@${REPLICA_DATABASE_HOST}:${REPLICA_DATABASE_PORT}/${REPLICA_DATABASE_DB}"

echo "Setting up Kasho database configuration..."

# Step 1: Verify prerequisites
echo "Verifying prerequisites..."
psql "$PRIMARY_SU_URL" -f sql/setup/verify-prerequisites.sql

# Step 2: Create Kasho user on both databases
echo "Creating Kasho user..."
# Process template with environment variables
./scripts/env-template-wrapper.sh < sql/setup/create-kasho-user.sql.template | psql "$PRIMARY_SU_URL"
./scripts/env-template-wrapper.sh < sql/setup/create-kasho-user.sql.template | psql "$REPLICA_SU_URL"

# Step 3: Set up DDL logging (before replication!)
echo "Setting up DDL logging..."
psql "$PRIMARY_SU_URL" -f sql/setup/setup-ddl-logging.sql

# Step 4: Set up replication
echo "Setting up replication..."
psql "$PRIMARY_SU_URL" -f sql/setup/setup-replication.sql

echo "Database setup complete!"
echo ""
echo "Next steps:"
echo "1. Ensure pg-change-stream is running"
echo "2. Run bootstrap-kasho.sh when ready to start replication"
```

To use this script:

```bash
# Set environment variables (from env.sample)
export PRIMARY_DATABASE_URL="postgresql://kasho:kasho@postgres-primary:5432/primary_db?sslmode=disable"
export REPLICA_DATABASE_URL="postgresql://kasho:kasho@postgres-replica:5432/replica_db?sslmode=disable"
export PRIMARY_DATABASE_SU_USER="postgres"
export PRIMARY_DATABASE_SU_PASSWORD="postgres"
export REPLICA_DATABASE_SU_USER="postgres"
export REPLICA_DATABASE_SU_PASSWORD="postgres"

# Run the setup script
./scripts/setup-kasho-db.sh

# Or with Docker
docker exec -it kasho ./scripts/setup-kasho-db.sh
```

### Manual Bootstrap Process

### Phase 1: Preparation

1. **DBA runs setup scripts** from `sql/setup/` to create users and configure replication
   - Prerequisites: WAL level configured in postgresql.conf, PostgreSQL restarted
   - Order: verify-prerequisites → create-kasho-user → setup-ddl-logging → setup-replication
2. **Ensure pg-change-stream is running** (will be in WAITING mode if no replication slot exists)

### Phase 2: Replication Slot Creation

The bootstrap coordinator creates or verifies the permanent replication slot:

1. **Create permanent replication slot**:
   ```sql
   SELECT slot_name, lsn
   FROM pg_create_logical_replication_slot('kasho_slot', 'pgoutput');
   ```
   
   **Why create the slot during bootstrap instead of setup?**
   
   - **Avoid WAL accumulation**: If created during setup but services aren't running, WAL files accumulate indefinitely
   - **Flexibility**: Allows custom slot names and multiple Kasho instances
   - **User permissions**: The kasho user can create slots (requires REPLICATION role), avoiding superuser requirement
   - **Clean recovery**: Easy to drop and recreate if needed
   
   **Why do we need a replication slot at all?**
   
   The replication slot serves as a **WAL retention guarantee**:
   - **Without a slot**: PostgreSQL might remove WAL files before pg-change-stream reads them
   - **With a slot**: PostgreSQL guarantees to keep all WAL from the slot's LSN forward
   - **During bootstrap**: Ensures no changes are lost while pg_dump runs (which can take hours)
   - **After bootstrap**: Prevents data loss if pg-change-stream is temporarily down
   
   The slot provides the critical coordination point - we know exactly which LSN to start from, and PostgreSQL guarantees all changes from that point are available.

2. **Signal pg-change-stream to start accumulating**:
   ```bash
   grpcurl -plaintext -d '{"start_lsn": "<lsn-from-step-1>"}' \
     localhost:50051 change_stream.ChangeStream/StartBootstrap
   ```
   
   Note: Ensure grpcurl is included in Docker images for bootstrap coordination.

### Phase 3: Data Transfer

1. **Take database dump**:
   ```bash
   pg_dump --no-owner --no-privileges source_db > dump.sql
   ```
   
   Note: pg_dump creates its own consistent snapshot internally. The replication slot ensures we don't lose any changes that occur during the dump.

2. **Start pg-bootstrap-sync** to process the dump:
   ```bash
   pg-bootstrap-sync --dump-file=dump.sql --redis-url=redis://localhost:6379 &
   ```
   
   This begins converting the SQL dump into change events. To enable parallel processing:
   
   **Option 1: Separate Redis Streams**
   - Bootstrap data → `kasho:bootstrap:changes` (ordered set)
   - WAL changes → `kasho:wal:changes` (ordered set)
   - pg-translicator reads from bootstrap first, then WAL
   
   **Option 2: Immediate Streaming with Markers**
   - Signal pg-change-stream to transition to STREAMING immediately
   - Use special marker events to coordinate:
     - `BOOTSTRAP_START` marker
     - Bootstrap changes (with fake LSNs < start_lsn)
     - `BOOTSTRAP_END` marker
     - Then accumulated WAL changes
   
   This allows pg-translicator to start applying bootstrap data immediately while pg-bootstrap-sync is still processing the dump.

### Phase 4: Monitor Completion

1. **Monitor bootstrap progress**:
   ```bash
   # Check if pg-bootstrap-sync is still running
   ps aux | grep pg-bootstrap-sync
   
   # Monitor Redis for bootstrap completion marker
   redis-cli GET "kasho:bootstrap:status"
   ```

2. **Verify all changes applied**:
   ```bash
   grpcurl -plaintext localhost:50051 kasho.ChangeStreamService/GetStatus
   ```

## Error Handling and Recovery

### Failed Bootstrap

If bootstrap fails at any stage:

1. **Stop pg-change-stream**
2. **Clear target database** using `sql/reset/reset-schema.sql`
3. **Clear Redis state**: Remove `kasho:change-stream:state` key
4. **Drop replication slots** if they exist
5. **Restart from Phase 1**

### State Machine Recovery

On pg-change-stream restart:

1. **Check Redis for saved state**
2. **If ACCUMULATING**: Resume from saved LSN
3. **If STREAMING**: Resume from replication slot position
4. **If no state**: Start in WAITING mode

### Connection Resilience

- Implement exponential backoff for connection retries
- Preserve state during temporary disconnections
- Log all state transitions with timestamps and reasons

## Monitoring and Observability

### Metrics

- `kasho_change_stream_state`: Current state (gauge)
- `kasho_change_stream_accumulated_total`: Total accumulated changes (counter)
- `kasho_change_stream_streamed_total`: Total streamed changes (counter)
- `kasho_change_stream_lag_bytes`: Replication lag in bytes (gauge)
- `kasho_bootstrap_duration_seconds`: Time spent in each phase (histogram)

### Health Checks

- `/health/ready`: Returns 200 only in STREAMING state
- `/health/live`: Returns 200 if process is running

### Logging

- State transitions with full context
- LSN progression during accumulation
- Client connections/disconnections
- Error conditions with remediation hints

## Implementation Timeline

### Phase 1: SQL Consolidation
- Reorganize SQL scripts into proposed structure (remove numeric prefixes from sql/ directory)
- Create symlinks with proper ordering for development environment
- Update documentation and deployment scripts

### Phase 2: Basic State Machine
- Implement three states in pg-change-stream
- Add Redis state persistence
- Create HTTP API endpoints

### Phase 3: Bootstrap Integration
- Add LSN-based accumulation logic
- Implement state transition logic
- Add comprehensive error handling

### Phase 4: Production Hardening
- Add monitoring and metrics
- Implement connection resilience
- Create operational runbooks
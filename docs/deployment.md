# Kasho Deployment Guide

Kasho is a data synchronization platform that consists of multiple cooperating services. While it has multiple components under the hood, we've designed deployment to be straightforward.

## Architecture Overview

Kasho consists of three main components:

1. **pg-change-stream**: Captures changes from PostgreSQL using logical replication
2. **pg-translicator**: Applies changes to the target database  
3. **Redis**: Provides message buffering between services

```
PostgreSQL Source → pg-change-stream → Redis → pg-translicator → Target Database
```

**Important:** Currently, only a single pg-translicator instance is supported. Multiple translicators would process the same changes multiple times.

## Configuration Requirements

### Environment Variables

#### pg-change-stream
| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `KV_URL` | Redis connection URL | Yes | `redis://redis:6379` |
| `PRIMARY_DATABASE_URL` | Source database connection URL | Yes | `postgresql://user:pass@host:5432/db?sslmode=disable` |

#### pg-translicator
| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `CHANGE_STREAM_SERVICE` | pg-change-stream service address | Yes | `pg-change-stream:8080` |
| `REPLICA_DATABASE_URL` | Target database connection URL | Yes | `postgresql://user:pass@host:5432/db?sslmode=disable` |

#### URL Format
Database URLs follow the standard PostgreSQL connection string format:
```
postgresql://[user[:password]@][host][:port][/database][?param=value&...]
```

Common parameters:
- `sslmode=disable|require|verify-ca|verify-full`
- `connect_timeout=30`
- `application_name=kasho`


### Configuration Files

pg-translicator requires a `transforms.yml` file mounted at `/app/config/transforms.yml`. This file defines table and column transformations.

For detailed information about the transforms.yml file format, available transform types, and configuration examples, see the [Transform Configuration Guide](transforms.md).

## Deployment

Since production deployment environments differ so significantly across different organizations, it's difficult to give perfect instructions for every environment. That said, in order to illustrate how to set up and use Kasho, here's how it could be done with Docker Compose.

### 1. Docker Compose Deployment

Create a `docker-compose.yml`:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  pg-change-stream:
    image: kasho:latest
    command: ./pg-change-stream
    environment:
      KV_URL: redis://redis:6379
      PRIMARY_DATABASE_URL: ${PRIMARY_DATABASE_URL}
    ports:
      - "8080:8080"
    depends_on:
      - redis

  pg-translicator:
    image: kasho:latest
    command: ./pg-translicator
    environment:
      CHANGE_STREAM_SERVICE: pg-change-stream:8080
      REPLICA_DATABASE_URL: ${REPLICA_DATABASE_URL}
    volumes:
      - ./config:/app/config:ro
    depends_on:
      - pg-change-stream

volumes:
  redis-data:
```

**Quick Start:**
```bash
# Create config directory with transforms.yml
mkdir -p config
cp environments/demo/config/transforms.yml config/

# Set environment variables in .env file
# Start all services
docker-compose up -d
```

## Database Requirements

### Source PostgreSQL
- Version 10+
- Logical replication enabled (`wal_level = logical`)
- Sufficient replication slots (`max_replication_slots` ≥ 2)
- User with REPLICATION permission and SELECT on all tables

### Target Database
- PostgreSQL 10+
- User with CREATE permission and full access to tables

### Initial Setup

Before deploying Kasho, the database must be configured:

1. **Configure PostgreSQL** (requires restart):
   ```ini
   # postgresql.conf
   wal_level = logical
   max_replication_slots = 10
   max_wal_senders = 10
   ```

2. **Run database setup** (requires superuser):
   ```bash
   # Set environment variables
   export PRIMARY_DATABASE_URL="postgresql://kasho:pass@primary:5432/db"
   export REPLICA_DATABASE_URL="postgresql://kasho:pass@replica:5432/db"
   export PRIMARY_DATABASE_SU_USER="postgres"
   export PRIMARY_DATABASE_SU_PASSWORD="postgres"
   export REPLICA_DATABASE_SU_USER="postgres"
   export REPLICA_DATABASE_SU_PASSWORD="postgres"
   
   # Run setup script
   ./scripts/prepare-primary-db.sh
   ```

   This script will:
   - Verify prerequisites (wal_level, etc.)
   - Create the kasho user with appropriate permissions
   - Set up DDL logging (if using DDL replication)
   - Create the publication (but NOT the replication slot - that's handled during bootstrap)

### Manual Database Setup Steps

If you prefer to set up the database manually instead of using the script:

1. **Verify WAL level** (as superuser on primary):
   ```sql
   -- Check current setting
   SHOW wal_level;
   
   -- If not 'logical', update postgresql.conf:
   -- wal_level = logical
   -- Then restart PostgreSQL
   ```

2. **Create Kasho user on primary** (as superuser):
   ```sql
   -- Create role with replication and login privileges
   CREATE ROLE kasho WITH REPLICATION LOGIN PASSWORD 'your-secure-password';
   
   -- Grant read permissions
   GRANT USAGE ON SCHEMA public TO kasho;
   GRANT CREATE ON SCHEMA public TO kasho;
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO kasho;
   GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO kasho;
   
   -- Grant future table permissions
   ALTER DEFAULT PRIVILEGES IN SCHEMA public 
   GRANT SELECT ON TABLES TO kasho;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public 
   GRANT SELECT ON SEQUENCES TO kasho;
   ```

3. **Create Kasho user on replica** (as superuser):
   ```sql
   -- Create role with same privileges
   CREATE ROLE kasho WITH REPLICATION LOGIN PASSWORD 'your-secure-password';
   
   -- Grant full permissions (needs to apply changes)
   GRANT USAGE ON SCHEMA public TO kasho;
   GRANT CREATE ON SCHEMA public TO kasho;
   GRANT ALL ON ALL TABLES IN SCHEMA public TO kasho;
   GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO kasho;
   
   -- Grant future table permissions
   ALTER DEFAULT PRIVILEGES IN SCHEMA public 
   GRANT ALL ON TABLES TO kasho;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public 
   GRANT ALL ON SEQUENCES TO kasho;
   ```

4. **Set up DDL logging** (as superuser on primary):
   ```sql
   -- Create DDL log table
   CREATE TABLE IF NOT EXISTS kasho_ddl_log (
       id BIGSERIAL PRIMARY KEY,
       lsn pg_lsn NOT NULL,
       timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       username TEXT NOT NULL,
       database TEXT NOT NULL,
       ddl TEXT NOT NULL
   );
   
   -- Create cleanup function
   CREATE OR REPLACE FUNCTION kasho_cleanup_ddl_log() RETURNS void AS $$
   BEGIN
       DELETE FROM kasho_ddl_log WHERE timestamp < NOW() - INTERVAL '7 days';
   END;
   $$ LANGUAGE plpgsql;
   
   -- Create DDL capture function
   CREATE OR REPLACE FUNCTION kasho_log_ddl() RETURNS event_trigger AS $$
   DECLARE
       current_lsn pg_lsn;
       ddl_text TEXT;
   BEGIN
       SELECT pg_current_wal_lsn() INTO current_lsn;
       SELECT current_query() INTO ddl_text;
       
       INSERT INTO kasho_ddl_log (lsn, username, database, ddl)
       VALUES (current_lsn, current_user, current_database(), ddl_text);
       
       PERFORM kasho_cleanup_ddl_log();
   END;
   $$ LANGUAGE plpgsql;
   
   -- Create event triggers
   CREATE EVENT TRIGGER kasho_log_ddl_start 
   ON ddl_command_start
   EXECUTE FUNCTION kasho_log_ddl();
   
   CREATE EVENT TRIGGER kasho_log_ddl_end 
   ON ddl_command_end
   EXECUTE FUNCTION kasho_log_ddl();
   ```

5. **Create publication** (as superuser on primary):
   ```sql
   -- Create publication for all tables
   CREATE PUBLICATION kasho_pub FOR ALL TABLES;
   ```

**Important:** Do NOT create the replication slot manually. It will be created automatically during the bootstrap process to ensure proper coordination between snapshot creation and change accumulation.

## Bootstrap Process

If you have existing data in your source database, you need to bootstrap it into Kasho:

### Understanding Bootstrap States

pg-change-stream operates in three states:
- **WAITING**: No replication slot exists, waiting for bootstrap to begin
- **ACCUMULATING**: Replication slot created, capturing changes during initial data load
- **STREAMING**: Normal operation, streaming all changes (both accumulated and new) to clients

The bootstrap process ensures no data loss by:
1. Creating a consistent snapshot of the source database
2. Starting change accumulation from that exact point
3. Loading the snapshot data
4. Transitioning to streaming mode with all accumulated changes

### Running Bootstrap

1. **Ensure pg-change-stream is running and in WAITING state**:
   ```bash
   grpcurl -plaintext pg-change-stream:8080 kasho.ChangeStreamService/GetStatus
   ```

2. **Run the bootstrap script**:
   ```bash
   # Basic usage - will prompt for confirmation before transitioning to streaming
   ./scripts/bootstrap-kasho.sh
   
   # Automatic mode - transitions to streaming without prompting
   WAIT_FOR_BOOTSTRAP=true ./scripts/bootstrap-kasho.sh
   ```

   This script will:
   - Create a temporary replication slot to get a consistent snapshot
   - Signal pg-change-stream to create its permanent slot and start accumulating changes
   - Take a database dump using the snapshot
   - Run pg-bootstrap-sync to convert the dump to change events
   - Clean up the temporary slot
   - Optionally transition to streaming mode (automatic with WAIT_FOR_BOOTSTRAP=true)

3. **Monitor progress**:
   ```bash
   # Check status during bootstrap
   grpcurl -plaintext pg-change-stream:8080 kasho.ChangeStreamService/GetStatus
   
   # If not using automatic mode, manually transition to streaming when ready
   grpcurl -plaintext pg-change-stream:8080 kasho.ChangeStreamService/CompleteBootstrap
   ```

### Manual Bootstrap Steps

If you prefer to run the bootstrap process manually:

1. **Create temporary snapshot and get LSN**:
   ```sql
   -- Create a temporary slot to get a consistent snapshot
   SELECT slot_name, lsn, snapshot_name 
   FROM pg_create_logical_replication_slot('kasho_temp_slot', 'pgoutput', true);
   ```

2. **Start accumulation** (this creates the permanent replication slot):
   ```bash
   grpcurl -plaintext -d '{"start_lsn": "<lsn>", "snapshot_name": "<snapshot>"}' \
     pg-change-stream:8080 kasho.ChangeStreamService/StartBootstrap
   ```

3. **Take dump and process**:
   ```bash
   # Use the snapshot from step 1 for consistency
   pg_dump --snapshot=<snapshot> --no-owner --no-privileges source_db > dump.sql
   
   # Convert dump to change events
   pg-bootstrap-sync --dump-file=dump.sql --redis-url=redis://redis:6379
   ```

4. **Clean up temporary slot and transition to streaming**:
   ```sql
   -- Drop the temporary slot (permanent slot remains)
   SELECT pg_drop_replication_slot('kasho_temp_slot');
   ```
   ```bash
   # Transition from ACCUMULATING to STREAMING state
   grpcurl -plaintext pg-change-stream:8080 kasho.ChangeStreamService/CompleteBootstrap
   ```

## Troubleshooting

### Services Won't Start

1. Check environment variables are set correctly
2. Verify database connectivity
3. Ensure transforms.yml is mounted for pg-translicator

### Connection Issues

Check that services can reach each other:
- pg-change-stream must be accessible on port 8080
- Redis must be accessible on its configured port
- Both databases must be reachable from their respective services

### Common Errors

**"PRIMARY_DATABASE_URL environment variable is required"**
- Set PRIMARY_DATABASE_URL in the format: `postgresql://user:password@host:port/database?sslmode=disable`
- Use `docker run --rm kasho` to see the help text with examples

**"Required config file /app/config/transforms.yml not found"**
- Mount the config directory with transforms.yml to /app/config

**"replication slot already exists"**
- This usually means a previous bootstrap wasn't cleaned up properly
- Check if pg-change-stream is in STREAMING state (bootstrap already complete)
- If stuck, manually drop the slot: `SELECT pg_drop_replication_slot('kasho_slot');`
- Then restart pg-change-stream to return to WAITING state
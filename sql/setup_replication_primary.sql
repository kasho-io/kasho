-- Run with: cat sql/setup_replication_primary.sql | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db
-- Note: After running this script, you can reset the databases using demo/scripts/reset-dbs.sh
-- Note: Demo SQL files (create_users_table.sql, add_dob_column_to_users_table.sql, insert_user.sql) are in demo/sql/
-- Note: WAL configuration is in setup_wal_primary.sql and must be run first

-- Drop and create replicator role
DROP ROLE IF EXISTS translicate;
CREATE ROLE translicate WITH REPLICATION LOGIN PASSWORD 'translicate';

-- Grant replication access
GRANT USAGE, CREATE ON SCHEMA public TO translicate;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO translicate;

-- Create publication for all tables and sequences
DROP PUBLICATION IF EXISTS translicate_pub;
CREATE PUBLICATION translicate_pub FOR ALL TABLES;

-- Create replication slot with wal2json
SELECT * FROM pg_create_logical_replication_slot('translicate_slot', 'wal2json');

CREATE TABLE IF NOT EXISTS translicate_ddl_log (
    id SERIAL PRIMARY KEY,
    lsn pg_lsn NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT now(),
    username TEXT,
    database TEXT,
    ddl TEXT NOT NULL
);

-- Function to clean up old entries
CREATE OR REPLACE FUNCTION cleanup_old_ddl_logs()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    -- Delete entries older than 7 days
    DELETE FROM translicate_ddl_log WHERE ts < NOW() - INTERVAL '7 days';
END;
$$;

-- Create a trigger to run cleanup after each insert
CREATE OR REPLACE FUNCTION trigger_cleanup_ddl_logs()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    -- Only run cleanup every 1000 inserts to avoid performance impact
    IF (SELECT count(*) FROM translicate_ddl_log) % 1000 = 0 THEN
        PERFORM cleanup_old_ddl_logs();
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER cleanup_ddl_logs_trigger
    AFTER INSERT ON translicate_ddl_log
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cleanup_ddl_logs();

CREATE OR REPLACE FUNCTION log_ddl_command()
RETURNS event_trigger
LANGUAGE plpgsql
AS $$
DECLARE
    ddl_stmt TEXT;
    current_lsn pg_lsn;
BEGIN
    SELECT pg_current_wal_lsn() INTO current_lsn;

    -- Get the full SQL statement
    SELECT current_setting('ddl.command', true) INTO ddl_stmt;

    INSERT INTO translicate_ddl_log(lsn, ddl, username, database)
    VALUES (current_lsn, ddl_stmt, SESSION_USER, current_database());
END;
$$;

-- Create a function to capture the DDL command
CREATE OR REPLACE FUNCTION capture_ddl_command()
RETURNS event_trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM set_config('ddl.command', current_query(), true);
END;
$$;

-- Create triggers to capture and log DDL
DROP EVENT TRIGGER IF EXISTS capture_ddl;
DROP EVENT TRIGGER IF EXISTS log_ddl;

CREATE EVENT TRIGGER capture_ddl
ON ddl_command_start
EXECUTE FUNCTION capture_ddl_command();

CREATE EVENT TRIGGER log_ddl
ON ddl_command_end
EXECUTE FUNCTION log_ddl_command();

-- Make sure translicate has access to everything
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO translicate;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO translicate;
GRANT SELECT ON translicate_ddl_log TO translicate;

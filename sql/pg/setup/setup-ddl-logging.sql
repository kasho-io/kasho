-- Setup DDL logging for Kasho (PostgreSQL version)
--
-- WHY THIS IS NEEDED FOR POSTGRESQL:
-- PostgreSQL's logical replication does not capture DDL (Data Definition Language)
-- statements like CREATE TABLE, ALTER TABLE, etc. Only DML (INSERT, UPDATE, DELETE)
-- is captured in the WAL stream. To replicate schema changes, we need to explicitly
-- capture DDL statements using PostgreSQL's event trigger system.
--
-- This script creates:
-- 1. kasho_ddl_log table - stores DDL statements with their WAL LSN position
-- 2. Event triggers - capture DDL on ddl_command_start and log on ddl_command_end
-- 3. Cleanup mechanism - removes entries older than 7 days
--
-- NOTE: MySQL does NOT need this file. MySQL's binary log automatically captures
-- both DDL and DML statements, so no explicit DDL logging mechanism is required.
-- The mysql-change-stream service reads DDL directly from the binlog.

DO $do_block$
BEGIN
  -- Drop and recreate kasho_ddl_log table
  IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'kasho_ddl_log') THEN
    DROP TABLE kasho_ddl_log;
  END IF;
  CREATE TABLE kasho_ddl_log (
    id SERIAL PRIMARY KEY,
    lsn pg_lsn NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT now(),
    username TEXT,
    database TEXT,
    ddl TEXT NOT NULL
  );

  -- Function to clean up old entries
  CREATE OR REPLACE FUNCTION kasho_cleanup_old_ddl_logs()
  RETURNS void
  LANGUAGE plpgsql
  AS $cleanup_func$
  BEGIN
    -- Delete entries older than 7 days
    DELETE FROM kasho_ddl_log WHERE ts < NOW() - INTERVAL '7 days';
  END;
  $cleanup_func$;

  -- Create a trigger to run cleanup after each insert
  CREATE OR REPLACE FUNCTION kasho_trigger_cleanup_ddl_logs()
  RETURNS trigger
  LANGUAGE plpgsql
  AS $trigger_func$
  BEGIN
    -- Only run cleanup every 1000 inserts to avoid performance impact
    IF (SELECT count(*) FROM kasho_ddl_log) % 1000 = 0 THEN
      PERFORM kasho_cleanup_old_ddl_logs();
    END IF;
    RETURN NEW;
  END;
  $trigger_func$;  

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'kasho_cleanup_ddl_logs_trigger') THEN
    DROP TRIGGER kasho_cleanup_ddl_logs_trigger ON kasho_ddl_log;
  END IF;
  CREATE TRIGGER kasho_cleanup_ddl_logs_trigger
    AFTER INSERT ON kasho_ddl_log
    FOR EACH ROW
    EXECUTE FUNCTION kasho_trigger_cleanup_ddl_logs();

  -- Create a function to log the DDL command
  CREATE OR REPLACE FUNCTION kasho_log_ddl_command()
  RETURNS event_trigger
  LANGUAGE plpgsql
  AS $log_func$
  DECLARE
    ddl_stmt TEXT;
    current_lsn pg_lsn;
  BEGIN
    SELECT pg_current_wal_lsn() INTO current_lsn;

    -- Get the full SQL statement
    SELECT current_setting('ddl.command', true) INTO ddl_stmt;

    INSERT INTO kasho_ddl_log(lsn, ddl, username, database)
    VALUES (current_lsn, ddl_stmt, SESSION_USER, current_database());
  END;
  $log_func$;

  -- Create a function to capture the DDL command
  CREATE OR REPLACE FUNCTION kasho_capture_ddl_command()
  RETURNS event_trigger
  LANGUAGE plpgsql
  AS $capture_func$
  BEGIN
    PERFORM set_config('ddl.command', current_query(), true);
  END;
  $capture_func$;

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_event_trigger WHERE evtname = 'kasho_capture_ddl') THEN
    DROP EVENT TRIGGER kasho_capture_ddl;
  END IF;
  CREATE EVENT TRIGGER kasho_capture_ddl
  ON ddl_command_start
  EXECUTE FUNCTION kasho_capture_ddl_command();

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_event_trigger WHERE evtname = 'kasho_log_ddl') THEN
    DROP EVENT TRIGGER kasho_log_ddl;
  END IF;
  CREATE EVENT TRIGGER kasho_log_ddl
  ON ddl_command_end
  EXECUTE FUNCTION kasho_log_ddl_command();
END;
$do_block$;

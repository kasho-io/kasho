DO $do_block$
BEGIN
  -- Drop and recreate translicate_ddl_log table
  IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'translicate_ddl_log') THEN
    DROP TABLE translicate_ddl_log;
  END IF;
  CREATE TABLE translicate_ddl_log (
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
  AS $cleanup_func$
  BEGIN
    -- Delete entries older than 7 days
    DELETE FROM translicate_ddl_log WHERE ts < NOW() - INTERVAL '7 days';
  END;
  $cleanup_func$;

  -- Create a trigger to run cleanup after each insert
  CREATE OR REPLACE FUNCTION trigger_cleanup_ddl_logs()
  RETURNS trigger
  LANGUAGE plpgsql
  AS $trigger_func$
  BEGIN
    -- Only run cleanup every 1000 inserts to avoid performance impact
    IF (SELECT count(*) FROM translicate_ddl_log) % 1000 = 0 THEN
      PERFORM cleanup_old_ddl_logs();
    END IF;
    RETURN NEW;
  END;
  $trigger_func$;  

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cleanup_ddl_logs_trigger') THEN
    DROP TRIGGER cleanup_ddl_logs_trigger ON translicate_ddl_log;
  END IF;
  CREATE TRIGGER cleanup_ddl_logs_trigger
    AFTER INSERT ON translicate_ddl_log
    FOR EACH ROW
    EXECUTE FUNCTION trigger_cleanup_ddl_logs();

  -- Create a function to log the DDL command
  CREATE OR REPLACE FUNCTION log_ddl_command()
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

    INSERT INTO translicate_ddl_log(lsn, ddl, username, database)
    VALUES (current_lsn, ddl_stmt, SESSION_USER, current_database());
  END;
  $log_func$;

  -- Create a function to capture the DDL command
  CREATE OR REPLACE FUNCTION capture_ddl_command()
  RETURNS event_trigger
  LANGUAGE plpgsql
  AS $capture_func$
  BEGIN
    PERFORM set_config('ddl.command', current_query(), true);
  END;
  $capture_func$;

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_event_trigger WHERE evtname = 'capture_ddl') THEN
    DROP EVENT TRIGGER capture_ddl;
  END IF;
  CREATE EVENT TRIGGER capture_ddl
  ON ddl_command_start
  EXECUTE FUNCTION capture_ddl_command();

  -- Drop and recreate event triggers to capture and log DDL
  IF EXISTS (SELECT 1 FROM pg_event_trigger WHERE evtname = 'log_ddl') THEN
    DROP EVENT TRIGGER log_ddl;
  END IF;
  CREATE EVENT TRIGGER log_ddl
  ON ddl_command_end
  EXECUTE FUNCTION log_ddl_command();
END;
$do_block$;

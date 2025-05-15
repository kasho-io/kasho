-- Drop replicator role if exists
DO $$
BEGIN
  -- Drop and create translicate role
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'translicate') THEN
    DROP ROLE translicate;
  END IF;
  CREATE ROLE translicate WITH REPLICATION LOGIN PASSWORD 'translicate';
  -- Grant necessary permissions
  GRANT USAGE, CREATE ON SCHEMA public TO translicate;
  GRANT SELECT ON ALL TABLES IN SCHEMA public TO translicate;
  GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO translicate;
  -- Set default privileges for future objects
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO translicate;
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO translicate;
END$$;
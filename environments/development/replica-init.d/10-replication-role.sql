-- Drop replicator role if exists
DO $$
BEGIN
  -- Drop and create kasho role
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'kasho') THEN
    DROP ROLE kasho;
  END IF;
  CREATE ROLE kasho WITH REPLICATION LOGIN PASSWORD 'kasho';
  -- Grant necessary permissions
  GRANT USAGE, CREATE ON SCHEMA public TO kasho;
  GRANT ALL ON ALL TABLES IN SCHEMA public TO kasho;
  GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO kasho;
  -- Set default privileges for future objects
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO kasho;
  ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO kasho; 
END$$; 



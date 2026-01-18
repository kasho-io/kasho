-- Create publication for Kasho replication
-- Note: This requires superuser privileges
DO $$
BEGIN
  -- Drop and recreate publication for all tables and sequences
  IF EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'kasho_pub') THEN
    DROP PUBLICATION kasho_pub;
  END IF;
  CREATE PUBLICATION kasho_pub FOR ALL TABLES;
END$$;

-- Note: Replication slot creation has been moved to the bootstrap process
-- This allows the kasho user to create the slot when needed, avoiding
-- WAL accumulation when services aren't running


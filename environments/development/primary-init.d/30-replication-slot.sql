-- First transaction: Handle publication
DO $$
BEGIN
  -- Drop and recreate publication for all tables and sequences
  IF EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'translicate_pub') THEN
    DROP PUBLICATION translicate_pub;
  END IF;
  CREATE PUBLICATION translicate_pub FOR ALL TABLES;
END$$;

-- Second transaction: Handle replication slot
DO $$
BEGIN
  -- Drop and recreate replication slot with pgoutput
  IF EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'translicate_slot') THEN
    PERFORM pg_drop_replication_slot('translicate_slot');
  END IF;
  PERFORM pg_create_logical_replication_slot('translicate_slot', 'pgoutput');
END$$;


#!/bin/bash
set -e

echo "🔄 Resetting demo-pg_primary-1..."
docker exec -i demo-pg_primary-1 psql -U postgres -d source_db <<'EOF'
-- Drop all tables in public schema
DO $$ 
DECLARE 
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;

-- Drop schema
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;

-- Drop slot if exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'translicate_slot') THEN
    PERFORM pg_drop_replication_slot('translicate_slot');
  END IF;
END$$;

-- Drop publication if exists
DROP PUBLICATION IF EXISTS translicate_pub;

-- Drop replicator role if exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'replicator') THEN
    DROP ROLE replicator;
  END IF;
END$$;
EOF

echo "🔄 Resetting demo-pg_replica-1..."
docker exec -i demo-pg_replica-1 psql -U postgres -d replica_db <<'EOF'
-- Drop all tables in public schema
DO $$ 
DECLARE 
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;

DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
EOF

echo "✅ Full reset complete."

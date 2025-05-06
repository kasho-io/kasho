-- Run with: cat sql/setup_replication_replica.sql | docker exec -i demo-pg_replica-1 psql -U postgres -d replica_db

-- Drop and create translicate role
DROP ROLE IF EXISTS translicate;
CREATE ROLE translicate WITH LOGIN PASSWORD 'translicate';

-- Grant necessary permissions
GRANT USAGE, CREATE ON SCHEMA public TO translicate;
GRANT ALL ON ALL TABLES IN SCHEMA public TO translicate;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO translicate;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO translicate;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO translicate; 
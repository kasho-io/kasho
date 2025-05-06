-- Run with: cat sql/setup_wal_primary.sql | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db
-- Note: These changes require a database restart to take effect
-- After running this script, restart the containers with: docker-compose down && docker-compose up -d

-- Enable logical replication
ALTER SYSTEM SET wal_level = logical;
ALTER SYSTEM SET max_wal_senders = 10;
ALTER SYSTEM SET max_replication_slots = 10; 
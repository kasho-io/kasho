-- Setup MySQL replication for Kasho
-- This configures the primary database for binary log replication

-- Verify GTID mode is enabled (recommended for Kasho)
SELECT @@gtid_mode AS gtid_mode;

-- Note: MySQL uses binary log-based replication
-- No explicit "publication" needed like PostgreSQL
-- The replication user needs REPLICATION SLAVE privilege
-- Binary logs will automatically capture all changes

-- To view current binary log position:
SHOW MASTER STATUS;

-- To view binary log events (for debugging):
-- SHOW BINLOG EVENTS IN 'mysql-bin.000001' LIMIT 10;

-- Verify MySQL prerequisites for Kasho replication
-- Run: mysql -u root -p < verify-prerequisites.sql

-- Check binary logging configuration
SELECT
    @@binlog_format AS binlog_format,
    @@log_bin AS log_bin_enabled,
    @@binlog_row_image AS binlog_row_image,
    @@server_id AS server_id;

-- Required settings:
-- binlog_format = ROW
-- log_bin = 1 (enabled)
-- binlog_row_image = FULL
-- server_id > 0 (unique per server)

-- To set these, add to my.cnf:
-- [mysqld]
-- server-id = 1
-- log_bin = mysql-bin
-- binlog_format = ROW
-- binlog_row_image = FULL
-- gtid_mode = ON
-- enforce_gtid_consistency = ON

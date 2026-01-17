-- Setup DDL logging for Kasho (MySQL version)
-- MySQL doesn't have event triggers, so we rely on binary log for DDL capture
-- This table stores DDL statements for reference

DROP TABLE IF EXISTS kasho_ddl_log;

CREATE TABLE kasho_ddl_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    position VARCHAR(255) NOT NULL,  -- GTID or binlog position
    logged_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6),
    username VARCHAR(255),
    database_name VARCHAR(255),
    ddl_text TEXT,
    INDEX idx_position (position),
    INDEX idx_logged_at (logged_at)
) ENGINE=InnoDB;

-- Cleanup procedure for old entries
DELIMITER //
CREATE PROCEDURE kasho_cleanup_ddl_log()
BEGIN
    DELETE FROM kasho_ddl_log
    WHERE logged_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
END//
DELIMITER ;

-- Create event to run cleanup daily (requires event_scheduler = ON)
DROP EVENT IF EXISTS kasho_ddl_cleanup_event;
CREATE EVENT kasho_ddl_cleanup_event
ON SCHEDULE EVERY 1 DAY
DO CALL kasho_cleanup_ddl_log();

-- Note: DDL statements are captured from binary log by mysql-change-stream
-- This table is populated by the application, not triggers
-- MySQL doesn't support DDL triggers like PostgreSQL's event triggers

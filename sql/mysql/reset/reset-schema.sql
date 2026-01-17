-- Reset MySQL database schema
-- WARNING: This will delete all data!

-- Disable foreign key checks for clean drop
SET FOREIGN_KEY_CHECKS = 0;

-- Get and drop all tables in current database
-- This procedure drops all tables dynamically
DELIMITER //
CREATE PROCEDURE drop_all_tables()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE tbl_name VARCHAR(255);
    DECLARE cur CURSOR FOR
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = DATABASE()
        AND table_type = 'BASE TABLE';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

    OPEN cur;

    read_loop: LOOP
        FETCH cur INTO tbl_name;
        IF done THEN
            LEAVE read_loop;
        END IF;
        SET @sql = CONCAT('DROP TABLE IF EXISTS `', tbl_name, '`');
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END LOOP;

    CLOSE cur;
END//
DELIMITER ;

CALL drop_all_tables();
DROP PROCEDURE IF EXISTS drop_all_tables;

-- Re-enable foreign key checks
SET FOREIGN_KEY_CHECKS = 1;

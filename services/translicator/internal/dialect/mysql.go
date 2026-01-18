package dialect

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"kasho/proto"
)

// MySQL implements the Dialect interface for MySQL databases
type MySQL struct{}

// NewMySQL creates a new MySQL dialect
func NewMySQL() *MySQL {
	return &MySQL{}
}

func (m *MySQL) Name() string {
	return "mysql"
}

func (m *MySQL) FormatValue(v *proto.ColumnValue) (string, error) {
	if v == nil || v.Value == nil {
		return "NULL", nil
	}

	switch val := v.Value.(type) {
	case *proto.ColumnValue_StringValue:
		// MySQL requires escaping backslashes as well as single quotes
		escaped := strings.ReplaceAll(val.StringValue, "'", "''")
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		return fmt.Sprintf("'%s'", escaped), nil
	case *proto.ColumnValue_IntValue:
		return fmt.Sprintf("%d", val.IntValue), nil
	case *proto.ColumnValue_FloatValue:
		return fmt.Sprintf("%f", val.FloatValue), nil
	case *proto.ColumnValue_BoolValue:
		// MySQL uses 1/0 for boolean values
		if val.BoolValue {
			return "1", nil
		}
		return "0", nil
	case *proto.ColumnValue_TimestampValue:
		// Try to parse as date first (YYYY-MM-DD)
		if t, err := time.Parse("2006-01-02", val.TimestampValue); err == nil {
			return fmt.Sprintf("'%s'", t.Format("2006-01-02")), nil
		}
		// Try to parse as timestamp
		if t, err := time.Parse(time.RFC3339, val.TimestampValue); err == nil {
			return fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05")), nil
		}
		return "", fmt.Errorf("invalid timestamp format: %s", val.TimestampValue)
	default:
		return "", fmt.Errorf("unsupported value type: %T", v.Value)
	}
}

func (m *MySQL) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
}

func (m *MySQL) SetupConnection(db *sql.DB) error {
	// Disable foreign key checks to allow replication without order dependencies
	_, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	return err
}

func (m *MySQL) SyncSequences(ctx context.Context, db *sql.DB) error {
	// MySQL uses AUTO_INCREMENT which is managed per-table
	// Query to find tables with auto_increment columns and sync their values
	query := `
		SELECT
			t.TABLE_SCHEMA,
			t.TABLE_NAME,
			c.COLUMN_NAME
		FROM information_schema.tables t
		JOIN information_schema.columns c
			ON c.TABLE_SCHEMA = t.TABLE_SCHEMA
			AND c.TABLE_NAME = t.TABLE_NAME
		WHERE t.TABLE_SCHEMA NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
		AND t.TABLE_TYPE = 'BASE TABLE'
		AND c.EXTRA LIKE '%auto_increment%'`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query auto_increment columns: %w", err)
	}
	defer rows.Close()

	updatedCount := 0
	for rows.Next() {
		var schema, table, column string
		if err := rows.Scan(&schema, &table, &column); err != nil {
			return fmt.Errorf("failed to scan auto_increment info: %w", err)
		}

		// Get max value for the column
		var maxVal sql.NullInt64
		maxQuery := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) + 1 FROM %s.%s",
			m.QuoteIdentifier(column), m.QuoteIdentifier(schema), m.QuoteIdentifier(table))
		err := db.QueryRowContext(ctx, maxQuery).Scan(&maxVal)
		if err != nil {
			log.Printf("Warning: failed to get max value for %s.%s.%s: %v", schema, table, column, err)
			continue
		}

		// Set the auto_increment value
		alterQuery := fmt.Sprintf("ALTER TABLE %s.%s AUTO_INCREMENT = %d",
			m.QuoteIdentifier(schema), m.QuoteIdentifier(table), maxVal.Int64)
		_, err = db.ExecContext(ctx, alterQuery)
		if err != nil {
			log.Printf("Warning: failed to set auto_increment for %s.%s: %v", schema, table, err)
			continue
		}
		updatedCount++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Updated %d auto_increment values", updatedCount)
	return nil
}

func (m *MySQL) GetUserTablesQuery() string {
	return `SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')
		AND table_type = 'BASE TABLE'`
}

func (m *MySQL) GetDriverName() string {
	return "mysql"
}

func (m *MySQL) FormatDSN(connStr string) string {
	// Convert URL format (mysql://user:pass@host:port/db) to MySQL DSN format (user:pass@tcp(host:port)/db)
	if !strings.HasPrefix(connStr, "mysql://") {
		// Already in DSN format or unknown format, return as-is
		return connStr
	}

	u, err := url.Parse(connStr)
	if err != nil {
		// Can't parse, return as-is
		return connStr
	}

	var dsn strings.Builder

	// Add user:password@
	if u.User != nil {
		dsn.WriteString(u.User.String())
		dsn.WriteString("@")
	}

	// Add tcp(host:port)
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "3306"
	}
	dsn.WriteString(fmt.Sprintf("tcp(%s:%s)", host, port))

	// Add /database
	dsn.WriteString(u.Path)

	// Add query parameters if any
	if u.RawQuery != "" {
		dsn.WriteString("?")
		dsn.WriteString(u.RawQuery)
	}

	return dsn.String()
}

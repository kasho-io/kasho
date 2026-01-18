package dialect

import (
	"context"
	"database/sql"
	"time"

	"kasho/proto"
)

// Dialect represents a SQL dialect for different database systems
type Dialect interface {
	// Name returns the dialect name (e.g., "postgresql", "mysql")
	Name() string

	// FormatValue formats a protobuf ColumnValue for SQL insertion
	FormatValue(val *proto.ColumnValue) (string, error)

	// QuoteIdentifier quotes a table or column name
	QuoteIdentifier(name string) string

	// SetupConnection runs any necessary setup commands on a new connection
	// e.g., PostgreSQL: SET session_replication_role = 'replica'
	// e.g., MySQL: SET FOREIGN_KEY_CHECKS = 0
	SetupConnection(db *sql.DB) error

	// SyncSequences synchronizes auto-increment/sequence values
	SyncSequences(ctx context.Context, db *sql.DB) error

	// GetUserTablesQuery returns SQL to count user tables
	GetUserTablesQuery() string

	// GetDriverName returns the database/sql driver name
	GetDriverName() string

	// FormatDSN converts a URL-style connection string to the driver's native DSN format
	FormatDSN(connStr string) string

	// Native type formatting methods (proto-free)
	// These can be used by tools that don't need protobuf dependencies

	// FormatString formats a string value for SQL, with proper escaping
	FormatString(s string) string

	// FormatInt formats an integer value for SQL
	FormatInt(i int64) string

	// FormatFloat formats a float value for SQL
	FormatFloat(f float64) string

	// FormatBool formats a boolean value for SQL
	// PostgreSQL uses true/false, MySQL uses 1/0
	FormatBool(b bool) string

	// FormatTimestamp formats a time.Time value for SQL
	FormatTimestamp(t time.Time) string

	// FormatDate formats a time.Time as a date-only value for SQL
	FormatDate(t time.Time) string

	// FormatNull returns the NULL literal for SQL
	FormatNull() string

	// DDL type methods for generating schema definitions

	// TypeUUID returns the column type for UUID values
	// PostgreSQL: "UUID", MySQL: "CHAR(36)"
	TypeUUID() string

	// TypeText returns the column type for unbounded text
	TypeText() string

	// TypeTimestamp returns the column type for timestamp with timezone
	// PostgreSQL: "TIMESTAMP WITH TIME ZONE", MySQL: "DATETIME(6)"
	TypeTimestamp() string

	// TypeDecimal returns the column type for decimal numbers
	TypeDecimal(precision, scale int) string

	// TypeInteger returns the column type for integers
	TypeInteger() string
}

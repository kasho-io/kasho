package dialect

import (
	"context"
	"database/sql"

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
}

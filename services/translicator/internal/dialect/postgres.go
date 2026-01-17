package dialect

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"kasho/proto"
)

// PostgreSQL implements the Dialect interface for PostgreSQL databases
type PostgreSQL struct{}

// NewPostgreSQL creates a new PostgreSQL dialect
func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

func (p *PostgreSQL) Name() string {
	return "postgresql"
}

func (p *PostgreSQL) FormatValue(v *proto.ColumnValue) (string, error) {
	if v == nil || v.Value == nil {
		return "NULL", nil
	}

	switch val := v.Value.(type) {
	case *proto.ColumnValue_StringValue:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val.StringValue, "'", "''")), nil
	case *proto.ColumnValue_IntValue:
		return fmt.Sprintf("%d", val.IntValue), nil
	case *proto.ColumnValue_FloatValue:
		return fmt.Sprintf("%f", val.FloatValue), nil
	case *proto.ColumnValue_BoolValue:
		return fmt.Sprintf("%t", val.BoolValue), nil
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

func (p *PostgreSQL) QuoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

// formatRegclass formats a schema.sequence name for use with setval()
// setval() takes a regclass, which can be a string literal
func (p *PostgreSQL) formatRegclass(schema, name string) string {
	// Escape single quotes in identifiers and wrap in quotes for regclass cast
	escapedSchema := strings.ReplaceAll(schema, "'", "''")
	escapedName := strings.ReplaceAll(name, "'", "''")
	return fmt.Sprintf("'%s.%s'", escapedSchema, escapedName)
}

func (p *PostgreSQL) SetupConnection(db *sql.DB) error {
	_, err := db.Exec("SET session_replication_role = 'replica'")
	return err
}

func (p *PostgreSQL) SyncSequences(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT
			n.nspname AS schema,
			t.relname AS table,
			a.attname AS column,
			s.relname AS sequence
		FROM pg_class s
		JOIN pg_depend d ON d.objid = s.oid AND d.deptype = 'a'
		JOIN pg_class t ON d.refobjid = t.oid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE s.relkind = 'S'`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query sequences: %w", err)
	}
	defer rows.Close()

	updatedCount := 0
	for rows.Next() {
		var schema, table, column, sequence string
		if err := rows.Scan(&schema, &table, &column, &sequence); err != nil {
			return fmt.Errorf("failed to scan sequence info: %w", err)
		}

		fullTable := fmt.Sprintf("%s.%s", p.QuoteIdentifier(schema), p.QuoteIdentifier(table))
		fullSeq := fmt.Sprintf("%s.%s", p.QuoteIdentifier(schema), p.QuoteIdentifier(sequence))

		var maxVal sql.NullInt64
		err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COALESCE(MAX(%s), 1) FROM %s", p.QuoteIdentifier(column), fullTable)).Scan(&maxVal)
		if err != nil {
			return fmt.Errorf("failed to get max value for %s.%s: %w", schema, table, err)
		}

		// Set the sequence to the max value (setval takes a regclass, so we pass the quoted name as a string)
		_, err = db.ExecContext(ctx, fmt.Sprintf("SELECT setval(%s, %d, true)", p.formatRegclass(schema, sequence), maxVal.Int64))
		if err != nil {
			return fmt.Errorf("failed to set sequence %s: %w", fullSeq, err)
		}
		updatedCount++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Updated %d sequences", updatedCount)
	return nil
}

func (p *PostgreSQL) GetUserTablesQuery() string {
	return `SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		AND table_type = 'BASE TABLE'`
}

func (p *PostgreSQL) GetDriverName() string {
	return "postgres"
}

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"pg-change-stream/api"
)

// ToSQL converts a Change into a SQL statement
func ToSQL(change *api.Change) (string, error) {
	switch data := change.Data.(type) {
	case *api.Change_Dml:
		return toDMLSQL(data.Dml)
	case *api.Change_Ddl:
		return data.Ddl.Ddl, nil
	default:
		return "", fmt.Errorf("unsupported change type: %T", change.Data)
	}
}

// toDMLSQL converts a DMLData into a SQL statement
func toDMLSQL(dml *api.DMLData) (string, error) {
	switch dml.Kind {
	case "insert":
		return toInsertSQL(dml)
	case "update":
		return toUpdateSQL(dml)
	case "delete":
		return toDeleteSQL(dml)
	default:
		return "", fmt.Errorf("unsupported DML kind: %s", dml.Kind)
	}
}

func formatValue(v string) string {
	// Handle NULL
	if v == "" || strings.ToLower(v) == "null" {
		return "NULL"
	}

	// Try to parse as timestamp
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
	}

	// Try to parse as date
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return fmt.Sprintf("'%s'", t.Format("2006-01-02"))
	}

	// Try to parse as number
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}

	// Try to parse as boolean
	if v == "true" || v == "false" {
		return v
	}

	// Otherwise treat as string
	return fmt.Sprintf("'%s'", v)
}

func toInsertSQL(dml *api.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	values := make([]string, len(dml.ColumnValues))
	for i, v := range dml.ColumnValues {
		values[i] = formatValue(v)
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		dml.Table,
		strings.Join(dml.ColumnNames, ", "),
		strings.Join(values, ", "),
	), nil
}

func toUpdateSQL(dml *api.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	if dml.OldKeys == nil {
		return "", fmt.Errorf("old keys required for update")
	}

	if len(dml.OldKeys.KeyNames) != len(dml.OldKeys.KeyValues) {
		return "", fmt.Errorf("mismatched old key names and values: %d names, %d values", len(dml.OldKeys.KeyNames), len(dml.OldKeys.KeyValues))
	}

	pkColumns := make(map[string]bool)
	for _, col := range dml.OldKeys.KeyNames {
		pkColumns[col] = true
	}

	setClauses := make([]string, 0)
	for i, col := range dml.ColumnNames {
		if !pkColumns[col] {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, formatValue(dml.ColumnValues[i])))
		}
	}

	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, col := range dml.OldKeys.KeyNames {
		whereClauses[i] = fmt.Sprintf("%s = %s", col, formatValue(dml.OldKeys.KeyValues[i]))
	}

	return fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s;",
		dml.Table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	), nil
}

func toDeleteSQL(dml *api.DMLData) (string, error) {
	if dml.OldKeys == nil {
		return "", fmt.Errorf("old keys required for delete")
	}

	if len(dml.OldKeys.KeyNames) != len(dml.OldKeys.KeyValues) {
		return "", fmt.Errorf("mismatched old key names and values: %d names, %d values", len(dml.OldKeys.KeyNames), len(dml.OldKeys.KeyValues))
	}

	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, col := range dml.OldKeys.KeyNames {
		whereClauses[i] = fmt.Sprintf("%s = %s", col, formatValue(dml.OldKeys.KeyValues[i]))
	}

	return fmt.Sprintf(
		"DELETE FROM %s WHERE %s;",
		dml.Table,
		strings.Join(whereClauses, " AND "),
	), nil
}

// SyncSequences synchronizes all sequences in the database to their corresponding table's max values
func SyncSequences(ctx context.Context, db *sql.DB) error {
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

		fullTable := fmt.Sprintf("%s.%s", schema, table)
		fullSeq := fmt.Sprintf("%s.%s", schema, sequence)

		var maxVal sql.NullInt64
		err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COALESCE(MAX(%s), 1) FROM %s", column, fullTable)).Scan(&maxVal)
		if err != nil {
			return fmt.Errorf("failed to get max value for %s: %w", fullTable, err)
		}

		// Set the sequence to the max value
		_, err = db.ExecContext(ctx, fmt.Sprintf("SELECT setval('%s', %d, true)", fullSeq, maxVal.Int64))
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

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"kasho/proto"
)

// ToSQL converts a Change into a SQL statement
func ToSQL(change *proto.Change) (string, error) {
	switch data := change.Data.(type) {
	case *proto.Change_Dml:
		return toDMLSQL(data.Dml)
	case *proto.Change_Ddl:
		return data.Ddl.Ddl, nil
	default:
		return "", fmt.Errorf("unsupported change type: %T", change.Data)
	}
}

// toDMLSQL converts a DMLData into a SQL statement
func toDMLSQL(dml *proto.DMLData) (string, error) {
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

// formatValue formats a value for SQL
func formatValue(v *proto.ColumnValue) (string, error) {
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

// toInsertSQL generates an INSERT SQL statement
func toInsertSQL(dml *proto.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	columns := strings.Join(dml.ColumnNames, ", ")
	values := make([]string, len(dml.ColumnValues))
	for i, v := range dml.ColumnValues {
		formatted, err := formatValue(v)
		if err != nil {
			return "", fmt.Errorf("error formatting value for column %s: %w", dml.ColumnNames[i], err)
		}
		values[i] = formatted
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", dml.Table, columns, strings.Join(values, ", ")), nil
}

// toUpdateSQL generates an UPDATE SQL statement
func toUpdateSQL(dml *proto.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	if dml.OldKeys == nil || len(dml.OldKeys.KeyNames) == 0 || len(dml.OldKeys.KeyValues) == 0 {
		return "", fmt.Errorf("update requires old keys")
	}

	// Build SET clause
	setClauses := make([]string, len(dml.ColumnNames))
	for i, col := range dml.ColumnNames {
		formatted, err := formatValue(dml.ColumnValues[i])
		if err != nil {
			return "", fmt.Errorf("error formatting value for column %s: %w", col, err)
		}
		setClauses[i] = fmt.Sprintf("%s = %s", col, formatted)
	}

	// Build WHERE clause
	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, key := range dml.OldKeys.KeyNames {
		formatted, err := formatValue(dml.OldKeys.KeyValues[i])
		if err != nil {
			return "", fmt.Errorf("error formatting value for key %s: %w", key, err)
		}
		whereClauses[i] = fmt.Sprintf("%s = %s", key, formatted)
	}

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s;",
		dml.Table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND ")), nil
}

// toDeleteSQL generates a DELETE SQL statement
func toDeleteSQL(dml *proto.DMLData) (string, error) {
	if dml.OldKeys == nil || len(dml.OldKeys.KeyNames) == 0 || len(dml.OldKeys.KeyValues) == 0 {
		return "", fmt.Errorf("delete requires old keys")
	}

	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, key := range dml.OldKeys.KeyNames {
		formatted, err := formatValue(dml.OldKeys.KeyValues[i])
		if err != nil {
			return "", fmt.Errorf("error formatting value for key %s: %w", key, err)
		}
		whereClauses[i] = fmt.Sprintf("%s = %s", key, formatted)
	}

	return fmt.Sprintf("DELETE FROM %s WHERE %s;",
		dml.Table,
		strings.Join(whereClauses, " AND ")), nil
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

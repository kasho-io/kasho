package sql

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"pg-change-stream/api"
)

// ToSQL converts a DMLData into a SQL statement
func ToSQL(dml *api.DMLData) (string, error) {
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

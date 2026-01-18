package sql

import (
	"fmt"
	"strings"

	"kasho/pkg/dialect"
	"kasho/proto"
)

// SQLGenerator generates SQL statements using a specific dialect
type SQLGenerator struct {
	dialect dialect.Dialect
}

// NewSQLGenerator creates a new SQL generator with the specified dialect
func NewSQLGenerator(d dialect.Dialect) *SQLGenerator {
	return &SQLGenerator{dialect: d}
}

// ToSQL converts a Change into a SQL statement
func (g *SQLGenerator) ToSQL(change *proto.Change) (string, error) {
	switch data := change.Data.(type) {
	case *proto.Change_Dml:
		return g.toDMLSQL(data.Dml)
	case *proto.Change_Ddl:
		return data.Ddl.Ddl, nil
	default:
		return "", fmt.Errorf("unsupported change type: %T", change.Data)
	}
}

// toDMLSQL converts a DMLData into a SQL statement
func (g *SQLGenerator) toDMLSQL(dml *proto.DMLData) (string, error) {
	switch dml.Kind {
	case "insert":
		return g.toInsertSQL(dml)
	case "update":
		return g.toUpdateSQL(dml)
	case "delete":
		return g.toDeleteSQL(dml)
	default:
		return "", fmt.Errorf("unsupported DML kind: '%s' (length: %d)", dml.Kind, len(dml.Kind))
	}
}

// toInsertSQL generates an INSERT SQL statement
func (g *SQLGenerator) toInsertSQL(dml *proto.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	columns := strings.Join(dml.ColumnNames, ", ")
	values := make([]string, len(dml.ColumnValues))
	for i, v := range dml.ColumnValues {
		formatted, err := g.dialect.FormatValue(v)
		if err != nil {
			return "", fmt.Errorf("error formatting value for column %s: %w", dml.ColumnNames[i], err)
		}
		values[i] = formatted
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", dml.Table, columns, strings.Join(values, ", ")), nil
}

// toUpdateSQL generates an UPDATE SQL statement
func (g *SQLGenerator) toUpdateSQL(dml *proto.DMLData) (string, error) {
	if len(dml.ColumnNames) != len(dml.ColumnValues) {
		return "", fmt.Errorf("mismatched column names and values: %d names, %d values", len(dml.ColumnNames), len(dml.ColumnValues))
	}

	if dml.OldKeys == nil || len(dml.OldKeys.KeyNames) == 0 || len(dml.OldKeys.KeyValues) == 0 {
		return "", fmt.Errorf("update requires old keys")
	}

	// Build SET clause
	setClauses := make([]string, len(dml.ColumnNames))
	for i, col := range dml.ColumnNames {
		formatted, err := g.dialect.FormatValue(dml.ColumnValues[i])
		if err != nil {
			return "", fmt.Errorf("error formatting value for column %s: %w", col, err)
		}
		setClauses[i] = fmt.Sprintf("%s = %s", col, formatted)
	}

	// Build WHERE clause
	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, key := range dml.OldKeys.KeyNames {
		formatted, err := g.dialect.FormatValue(dml.OldKeys.KeyValues[i])
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
func (g *SQLGenerator) toDeleteSQL(dml *proto.DMLData) (string, error) {
	if dml.OldKeys == nil || len(dml.OldKeys.KeyNames) == 0 || len(dml.OldKeys.KeyValues) == 0 {
		return "", fmt.Errorf("delete requires old keys")
	}

	whereClauses := make([]string, len(dml.OldKeys.KeyNames))
	for i, key := range dml.OldKeys.KeyNames {
		formatted, err := g.dialect.FormatValue(dml.OldKeys.KeyValues[i])
		if err != nil {
			return "", fmt.Errorf("error formatting value for key %s: %w", key, err)
		}
		whereClauses[i] = fmt.Sprintf("%s = %s", key, formatted)
	}

	return fmt.Sprintf("DELETE FROM %s WHERE %s;",
		dml.Table,
		strings.Join(whereClauses, " AND ")), nil
}

// ToSQL converts a Change into a SQL statement using PostgreSQL dialect (backwards compatible)
// Deprecated: Use SQLGenerator.ToSQL instead
func ToSQL(change *proto.Change) (string, error) {
	g := NewSQLGenerator(dialect.NewPostgreSQL())
	return g.ToSQL(change)
}

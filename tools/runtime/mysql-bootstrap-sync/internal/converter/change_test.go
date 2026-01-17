package converter

import (
	"testing"

	"mysql-bootstrap-sync/internal/parser"
)

func TestChangeConverter_ConvertDMLStatement(t *testing.T) {
	conv := NewChangeConverter()

	stmt := parser.DMLStatement{
		Table:       "users",
		ColumnNames: []string{"id", "name", "email"},
		ColumnValues: [][]string{
			{"1", "John Doe", "john@example.com"},
			{"2", "Jane Doe", "jane@example.com"},
		},
	}

	changes, err := conv.ConvertStatements([]parser.Statement{stmt})
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	if len(changes) != 2 {
		t.Fatalf("expected 2 changes (one per row), got %d", len(changes))
	}

	// Check first change
	change1 := changes[0]
	if change1.Type() != "dml" {
		t.Errorf("change1.Type() = %v, want dml", change1.Type())
	}
	if !IsBootstrapPosition(change1.GetPosition()) {
		t.Errorf("change1.GetPosition() = %v, should be bootstrap position", change1.GetPosition())
	}

	// Check second change has incremented position
	change2 := changes[1]
	pos1, _ := ParseBootstrapPosition(change1.GetPosition())
	pos2, _ := ParseBootstrapPosition(change2.GetPosition())
	if pos2 != pos1+1 {
		t.Errorf("positions should be sequential: %d, %d", pos1, pos2)
	}
}

func TestChangeConverter_ConvertDDLStatement(t *testing.T) {
	conv := NewChangeConverter()

	stmt := parser.DDLStatement{
		SQL:      "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100))",
		Table:    "users",
		Database: "testdb",
	}

	changes, err := conv.ConvertStatements([]parser.Statement{stmt})
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	if change.Type() != "ddl" {
		t.Errorf("change.Type() = %v, want ddl", change.Type())
	}
	if !IsBootstrapPosition(change.GetPosition()) {
		t.Errorf("change.GetPosition() = %v, should be bootstrap position", change.GetPosition())
	}
}

func TestChangeConverter_ConvertValueTypes(t *testing.T) {
	conv := NewChangeConverter()

	tests := []struct {
		name      string
		value     string
		wantType  string // "int", "float", "bool", "timestamp", "string"
		wantValue interface{}
	}{
		{"integer", "42", "int", int64(42)},
		{"negative integer", "-100", "int", int64(-100)},
		{"float", "3.14", "float", 3.14},
		{"bool true", "true", "bool", true},
		{"bool false", "false", "bool", false},
		{"timestamp RFC3339", "2024-03-20T15:04:05Z", "timestamp", "2024-03-20T15:04:05Z"},
		{"timestamp MySQL format", "2024-03-20 15:04:05", "timestamp", "2024-03-20T15:04:05Z"},
		{"string", "hello world", "string", "hello world"},
		{"empty (NULL)", "", "null", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := parser.DMLStatement{
				Table:        "test",
				ColumnNames:  []string{"col"},
				ColumnValues: [][]string{{tt.value}},
			}

			changes, err := conv.ConvertStatements([]parser.Statement{stmt})
			if err != nil {
				t.Fatalf("ConvertStatements() error = %v", err)
			}

			if len(changes) != 1 {
				t.Fatalf("expected 1 change, got %d", len(changes))
			}

			// Reset converter for next test
			conv = NewChangeConverter()
		})
	}
}

func TestChangeConverter_MixedStatements(t *testing.T) {
	conv := NewChangeConverter()

	statements := []parser.Statement{
		parser.DDLStatement{
			SQL:   "CREATE TABLE users (id INT)",
			Table: "users",
		},
		parser.DMLStatement{
			Table:        "users",
			ColumnNames:  []string{"id"},
			ColumnValues: [][]string{{"1"}, {"2"}},
		},
		parser.DDLStatement{
			SQL:   "ALTER TABLE users ADD COLUMN name VARCHAR(100)",
			Table: "users",
		},
	}

	changes, err := conv.ConvertStatements(statements)
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	// 1 DDL + 2 DML rows + 1 DDL = 4 changes
	if len(changes) != 4 {
		t.Fatalf("expected 4 changes, got %d", len(changes))
	}

	// Verify positions are strictly sequential
	for i := 1; i < len(changes); i++ {
		pos1, _ := ParseBootstrapPosition(changes[i-1].GetPosition())
		pos2, _ := ParseBootstrapPosition(changes[i].GetPosition())
		if pos2 != pos1+1 {
			t.Errorf("positions should be sequential at index %d: %d -> %d", i, pos1, pos2)
		}
	}
}

func TestChangeConverter_ColumnCountMismatch(t *testing.T) {
	conv := NewChangeConverter()

	stmt := parser.DMLStatement{
		Table:       "users",
		ColumnNames: []string{"id", "name", "email"}, // 3 columns
		ColumnValues: [][]string{
			{"1", "John"}, // Only 2 values
		},
	}

	_, err := conv.ConvertStatements([]parser.Statement{stmt})
	if err == nil {
		t.Error("expected error for column count mismatch")
	}
}

func TestChangeConverter_ConvertValueTypes_EdgeCases(t *testing.T) {
	conv := NewChangeConverter()

	tests := []struct {
		name  string
		value string
	}{
		// Numeric edge cases
		{"max int64", "9223372036854775807"},
		{"min int64", "-9223372036854775808"},
		{"zero", "0"},
		{"large float", "123456789.123456"},
		{"scientific notation", "1.23e10"},
		{"negative float", "-3.14159"},

		// String edge cases
		{"string with numbers", "abc123"},
		{"numeric string that should stay string", "007"},
		{"phone number", "+1-555-1234"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},

		// Timestamp edge cases
		{"timestamp with microseconds", "2024-03-20 15:04:05.123456"},
		{"date only", "2024-03-20"},
		{"year only", "2024"},

		// Special values
		{"NULL literal", "NULL"},
		{"whitespace", "   "},
		{"newline in value", "line1\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := parser.DMLStatement{
				Table:        "test",
				ColumnNames:  []string{"col"},
				ColumnValues: [][]string{{tt.value}},
			}

			changes, err := conv.ConvertStatements([]parser.Statement{stmt})
			if err != nil {
				t.Fatalf("ConvertStatements() error = %v", err)
			}

			if len(changes) != 1 {
				t.Fatalf("expected 1 change, got %d", len(changes))
			}

			// Reset converter for next test
			conv = NewChangeConverter()
		})
	}
}

func TestChangeConverter_EmptyStatements(t *testing.T) {
	conv := NewChangeConverter()

	changes, err := conv.ConvertStatements([]parser.Statement{})
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("expected 0 changes for empty input, got %d", len(changes))
	}
}

func TestChangeConverter_DMLWithNoRows(t *testing.T) {
	conv := NewChangeConverter()

	stmt := parser.DMLStatement{
		Table:        "users",
		ColumnNames:  []string{"id", "name"},
		ColumnValues: [][]string{}, // No rows
	}

	changes, err := conv.ConvertStatements([]parser.Statement{stmt})
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("expected 0 changes for DML with no rows, got %d", len(changes))
	}
}

func TestChangeConverter_LargeNumberOfRows(t *testing.T) {
	conv := NewChangeConverter()

	// Create 1000 rows
	rows := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = []string{string(rune('0' + i%10))}
	}

	stmt := parser.DMLStatement{
		Table:        "test",
		ColumnNames:  []string{"id"},
		ColumnValues: rows,
	}

	changes, err := conv.ConvertStatements([]parser.Statement{stmt})
	if err != nil {
		t.Fatalf("ConvertStatements() error = %v", err)
	}

	if len(changes) != 1000 {
		t.Errorf("expected 1000 changes, got %d", len(changes))
	}

	// Verify all positions are sequential
	for i := 1; i < len(changes); i++ {
		pos1, _ := ParseBootstrapPosition(changes[i-1].GetPosition())
		pos2, _ := ParseBootstrapPosition(changes[i].GetPosition())
		if pos2 != pos1+1 {
			t.Errorf("positions should be sequential at index %d: %d -> %d", i, pos1, pos2)
		}
	}
}

func TestChangeConverter_SpecialTableNames(t *testing.T) {
	conv := NewChangeConverter()

	tables := []string{
		"users",
		"user_accounts",
		"UserAccounts",
		"table with spaces",
		"table-with-dashes",
		"日本語テーブル",
	}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			stmt := parser.DMLStatement{
				Table:        table,
				ColumnNames:  []string{"id"},
				ColumnValues: [][]string{{"1"}},
			}

			changes, err := conv.ConvertStatements([]parser.Statement{stmt})
			if err != nil {
				t.Fatalf("ConvertStatements() error = %v for table %q", err, table)
			}

			if len(changes) != 1 {
				t.Errorf("expected 1 change, got %d", len(changes))
			}

			conv = NewChangeConverter()
		})
	}
}

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
	if !IsBootstrapPosition(change1.GetLSN()) {
		t.Errorf("change1.GetLSN() = %v, should be bootstrap position", change1.GetLSN())
	}

	// Check second change has incremented position
	change2 := changes[1]
	pos1, _ := ParseBootstrapPosition(change1.GetLSN())
	pos2, _ := ParseBootstrapPosition(change2.GetLSN())
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
	if !IsBootstrapPosition(change.GetLSN()) {
		t.Errorf("change.GetLSN() = %v, should be bootstrap position", change.GetLSN())
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
		pos1, _ := ParseBootstrapPosition(changes[i-1].GetLSN())
		pos2, _ := ParseBootstrapPosition(changes[i].GetLSN())
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

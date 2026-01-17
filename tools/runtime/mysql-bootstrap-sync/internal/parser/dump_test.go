package parser

import (
	"strings"
	"testing"
)

func TestDumpParser_ParseInsertStatement(t *testing.T) {
	parser := NewDumpParser()

	tests := []struct {
		name           string
		sql            string
		wantTable      string
		wantCols       []string
		wantRowCount   int
		wantFirstValue string
	}{
		{
			name:           "simple insert with columns",
			sql:            "INSERT INTO `users` (`id`, `name`, `email`) VALUES (1, 'John Doe', 'john@example.com');",
			wantTable:      "users",
			wantCols:       []string{"id", "name", "email"},
			wantRowCount:   1,
			wantFirstValue: "1",
		},
		{
			name:           "extended insert multiple rows",
			sql:            "INSERT INTO `users` (`id`, `name`) VALUES (1, 'John'), (2, 'Jane'), (3, 'Bob');",
			wantTable:      "users",
			wantCols:       []string{"id", "name"},
			wantRowCount:   3,
			wantFirstValue: "1",
		},
		{
			name:           "insert with NULL values",
			sql:            "INSERT INTO `users` (`id`, `name`, `email`) VALUES (1, 'John', NULL);",
			wantTable:      "users",
			wantCols:       []string{"id", "name", "email"},
			wantRowCount:   1,
			wantFirstValue: "1",
		},
		{
			name:           "insert without column names",
			sql:            "INSERT INTO users VALUES (1, 'John', 'john@example.com');",
			wantTable:      "users",
			wantCols:       nil,
			wantRowCount:   1,
			wantFirstValue: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseStream(strings.NewReader(tt.sql))
			if err != nil {
				t.Fatalf("ParseStream() error = %v", err)
			}

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			dml, ok := result.Statements[0].(DMLStatement)
			if !ok {
				t.Fatalf("expected DMLStatement, got %T", result.Statements[0])
			}

			if dml.Table != tt.wantTable {
				t.Errorf("table = %v, want %v", dml.Table, tt.wantTable)
			}

			if tt.wantCols != nil {
				if len(dml.ColumnNames) != len(tt.wantCols) {
					t.Errorf("column count = %d, want %d", len(dml.ColumnNames), len(tt.wantCols))
				}
			}

			if len(dml.ColumnValues) != tt.wantRowCount {
				t.Errorf("row count = %d, want %d", len(dml.ColumnValues), tt.wantRowCount)
			}

			if len(dml.ColumnValues) > 0 && len(dml.ColumnValues[0]) > 0 {
				if dml.ColumnValues[0][0] != tt.wantFirstValue {
					t.Errorf("first value = %v, want %v", dml.ColumnValues[0][0], tt.wantFirstValue)
				}
			}
		})
	}
}

func TestDumpParser_ParseDDLStatement(t *testing.T) {
	parser := NewDumpParser()

	tests := []struct {
		name      string
		sql       string
		wantTable string
	}{
		{
			name:      "create table",
			sql:       "CREATE TABLE `users` (\n  `id` int NOT NULL AUTO_INCREMENT,\n  `name` varchar(100) DEFAULT NULL,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB;",
			wantTable: "users",
		},
		{
			name:      "alter table",
			sql:       "ALTER TABLE `users` ADD COLUMN `email` varchar(255);",
			wantTable: "users",
		},
		{
			name:      "drop table",
			sql:       "DROP TABLE IF EXISTS `old_table`;",
			wantTable: "old_table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseStream(strings.NewReader(tt.sql))
			if err != nil {
				t.Fatalf("ParseStream() error = %v", err)
			}

			if len(result.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(result.Statements))
			}

			ddl, ok := result.Statements[0].(DDLStatement)
			if !ok {
				t.Fatalf("expected DDLStatement, got %T", result.Statements[0])
			}

			if ddl.Table != tt.wantTable {
				t.Errorf("table = %v, want %v", ddl.Table, tt.wantTable)
			}
		})
	}
}

func TestDumpParser_SkipsKashoTables(t *testing.T) {
	parser := NewDumpParser()

	sql := `
CREATE TABLE kasho_ddl_log (id INT);
INSERT INTO kasho_ddl_log (id) VALUES (1);
CREATE TABLE users (id INT);
INSERT INTO users (id) VALUES (1);
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	// Should only have 2 statements (the users table, not kasho_ddl_log)
	if len(result.Statements) != 2 {
		t.Errorf("expected 2 statements (skipping kasho_*), got %d", len(result.Statements))
	}

	// Verify it's the users table
	for _, stmt := range result.Statements {
		switch s := stmt.(type) {
		case DDLStatement:
			if strings.Contains(s.Table, "kasho_") {
				t.Errorf("should have skipped kasho table, got: %s", s.Table)
			}
		case DMLStatement:
			if strings.Contains(s.Table, "kasho_") {
				t.Errorf("should have skipped kasho table, got: %s", s.Table)
			}
		}
	}
}

func TestDumpParser_SkipsSessionCommands(t *testing.T) {
	parser := NewDumpParser()

	sql := `
SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT;
SET NAMES utf8mb4;
LOCK TABLES users WRITE;
INSERT INTO users (id) VALUES (1);
UNLOCK TABLES;
COMMIT;
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	// Should only have 1 statement (the INSERT)
	if len(result.Statements) != 1 {
		t.Errorf("expected 1 statement (skipping session commands), got %d", len(result.Statements))
	}
}

func TestDumpParser_ExtendedInsertWithEscapes(t *testing.T) {
	parser := NewDumpParser()

	sql := `INSERT INTO users (id, name, bio) VALUES (1, 'John O\'Brien', 'Line 1\nLine 2');`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	dml := result.Statements[0].(DMLStatement)
	if len(dml.ColumnValues) != 1 {
		t.Fatalf("expected 1 row, got %d", len(dml.ColumnValues))
	}

	// Check that escapes are properly handled
	name := dml.ColumnValues[0][1]
	if name != "John O'Brien" {
		t.Errorf("name = %v, want 'John O'Brien'", name)
	}

	bio := dml.ColumnValues[0][2]
	if bio != "Line 1\nLine 2" {
		t.Errorf("bio = %v, want 'Line 1\\nLine 2'", bio)
	}
}

func TestDumpParser_MaxRowsPerTable(t *testing.T) {
	parser := NewDumpParser()
	parser.MaxRowsPerTable = 2

	sql := `INSERT INTO users (id) VALUES (1), (2), (3), (4), (5);`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	dml := result.Statements[0].(DMLStatement)
	if len(dml.ColumnValues) != 2 {
		t.Errorf("expected 2 rows (limited), got %d", len(dml.ColumnValues))
	}
}

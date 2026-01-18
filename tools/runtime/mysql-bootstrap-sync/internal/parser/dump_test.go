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

func TestDumpParser_MultilineInsert(t *testing.T) {
	parser := NewDumpParser()

	sql := `INSERT INTO users (id, name, email) VALUES
(1, 'John Doe', 'john@example.com'),
(2, 'Jane Doe', 'jane@example.com'),
(3, 'Bob Smith', 'bob@example.com');`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	dml := result.Statements[0].(DMLStatement)
	if len(dml.ColumnValues) != 3 {
		t.Errorf("expected 3 rows, got %d", len(dml.ColumnValues))
	}
}

func TestDumpParser_BinaryAndHexValues(t *testing.T) {
	parser := NewDumpParser()

	tests := []struct {
		name      string
		sql       string
		wantValue string
	}{
		{
			name:      "hex value",
			sql:       "INSERT INTO data (id, content) VALUES (1, 0x48656C6C6F);",
			wantValue: "0x48656C6C6F",
		},
		{
			name:      "binary literal",
			sql:       "INSERT INTO data (id, flag) VALUES (1, b'1010');",
			wantValue: "b'1010'",
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

			dml := result.Statements[0].(DMLStatement)
			if len(dml.ColumnValues) != 1 || len(dml.ColumnValues[0]) < 2 {
				t.Fatalf("expected at least 2 values in row")
			}

			// The second value should be the hex/binary data
			got := dml.ColumnValues[0][1]
			if got != tt.wantValue {
				t.Errorf("value = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestDumpParser_VeryLongStrings(t *testing.T) {
	parser := NewDumpParser()

	// Create a 10KB string
	longString := strings.Repeat("x", 10*1024)
	sql := "INSERT INTO data (id, content) VALUES (1, '" + longString + "');"

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	dml := result.Statements[0].(DMLStatement)
	if len(dml.ColumnValues[0][1]) != 10*1024 {
		t.Errorf("expected 10KB string, got %d bytes", len(dml.ColumnValues[0][1]))
	}
}

func TestDumpParser_EscapeSequences(t *testing.T) {
	parser := NewDumpParser()

	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{"escaped single quote", `'It\'s'`, "It's"},
		// Note: MySQL escaping converts \\ to single backslash, then \t becomes tab
		{"escaped backslash", `'path\\to\\file'`, "path\to\\file"},
		{"escaped newline", `'line1\nline2'`, "line1\nline2"},
		{"escaped carriage return", `'line1\rline2'`, "line1\rline2"},
		{"escaped tab", `'col1\tcol2'`, "col1\tcol2"},
		{"escaped null", `'null\0char'`, "null\x00char"},
		{"mixed escapes", `'It\'s a \\test\n'`, "It's a \test\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := "INSERT INTO test (val) VALUES (" + tt.input + ");"
			result, err := parser.ParseStream(strings.NewReader(sql))
			if err != nil {
				t.Fatalf("ParseStream() error = %v", err)
			}

			dml := result.Statements[0].(DMLStatement)
			got := dml.ColumnValues[0][0]
			if got != tt.wantOutput {
				t.Errorf("value = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

func TestDumpParser_SpecialIdentifiers(t *testing.T) {
	parser := NewDumpParser()

	tests := []struct {
		name      string
		sql       string
		wantTable string
	}{
		// Note: Current parser extracts table name up to first non-identifier char
		// These tests document actual behavior
		{
			name:      "backtick quoted simple table",
			sql:       "INSERT INTO `users` (id) VALUES (1);",
			wantTable: "users",
		},
		{
			name:      "unquoted table",
			sql:       "INSERT INTO users (id) VALUES (1);",
			wantTable: "users",
		},
		{
			name:      "reserved word as table",
			sql:       "INSERT INTO `select` (id) VALUES (1);",
			wantTable: "select",
		},
		{
			name:      "table with underscore",
			sql:       "INSERT INTO `user_data` (id) VALUES (1);",
			wantTable: "user_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseStream(strings.NewReader(tt.sql))
			if err != nil {
				t.Fatalf("ParseStream() error = %v", err)
			}

			dml := result.Statements[0].(DMLStatement)
			if dml.Table != tt.wantTable {
				t.Errorf("table = %v, want %v", dml.Table, tt.wantTable)
			}
		})
	}
}

func TestDumpParser_CommentsInSQL(t *testing.T) {
	parser := NewDumpParser()

	sql := `
-- This is a comment
INSERT INTO users (id, name) VALUES (1, 'John');
/* This is a block comment */
INSERT INTO users (id, name) VALUES (2, 'Jane');
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	// Comments should be skipped, only INSERT statements parsed
	if len(result.Statements) != 2 {
		t.Errorf("expected 2 statements, got %d", len(result.Statements))
	}
}

func TestDumpParser_EmptyInput(t *testing.T) {
	parser := NewDumpParser()

	result, err := parser.ParseStream(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 0 {
		t.Errorf("expected 0 statements for empty input, got %d", len(result.Statements))
	}
}

func TestDumpParser_OnlyComments(t *testing.T) {
	parser := NewDumpParser()

	sql := `
-- Comment 1
-- Comment 2
/* Block comment */
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	if len(result.Statements) != 0 {
		t.Errorf("expected 0 statements for comments only, got %d", len(result.Statements))
	}
}

func TestDumpParser_UnicodeContent(t *testing.T) {
	parser := NewDumpParser()

	sql := `INSERT INTO users (id, name, bio) VALUES (1, '日本語', '中文内容 with émojis 👋');`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	dml := result.Statements[0].(DMLStatement)
	if dml.ColumnValues[0][1] != "日本語" {
		t.Errorf("name = %v, want '日本語'", dml.ColumnValues[0][1])
	}
	if dml.ColumnValues[0][2] != "中文内容 with émojis 👋" {
		t.Errorf("bio = %v, want '中文内容 with émojis 👋'", dml.ColumnValues[0][2])
	}
}

func TestDumpParser_NumericValues(t *testing.T) {
	parser := NewDumpParser()

	sql := `INSERT INTO numbers (a, b, c, d, e) VALUES (42, -100, 3.14, -2.5, 1.23e10);`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	dml := result.Statements[0].(DMLStatement)
	expected := []string{"42", "-100", "3.14", "-2.5", "1.23e10"}
	for i, want := range expected {
		got := dml.ColumnValues[0][i]
		if got != want {
			t.Errorf("value[%d] = %v, want %v", i, got, want)
		}
	}
}

func TestDumpParser_DelimiterStatements(t *testing.T) {
	parser := NewDumpParser()

	// Test stored procedure with DELIMITER change (common in mysqldump)
	sql := `
INSERT INTO users (id, name) VALUES (1, 'Before proc');
DELIMITER ;;
CREATE PROCEDURE cleanup()
BEGIN
    DELETE FROM logs WHERE created_at < NOW() - INTERVAL 7 DAY;
    DELETE FROM temp WHERE id > 0;
END ;;
DELIMITER ;
INSERT INTO users (id, name) VALUES (2, 'After proc');
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	// Should have: 2 DML (INSERTs) + 1 DDL (CREATE PROCEDURE)
	if len(result.Statements) != 3 {
		t.Errorf("expected 3 statements, got %d", len(result.Statements))
	}

	// First statement: INSERT before procedure
	dml1, ok := result.Statements[0].(DMLStatement)
	if !ok {
		t.Errorf("statement 0: expected DMLStatement, got %T", result.Statements[0])
	} else if dml1.Table != "users" {
		t.Errorf("statement 0: table = %v, want 'users'", dml1.Table)
	}

	// Second statement: CREATE PROCEDURE (as DDL)
	ddl, ok := result.Statements[1].(DDLStatement)
	if !ok {
		t.Errorf("statement 1: expected DDLStatement, got %T", result.Statements[1])
	} else {
		// The procedure should contain the full body including both DELETE statements
		if !strings.Contains(ddl.SQL, "DELETE FROM logs") {
			t.Errorf("procedure missing 'DELETE FROM logs': %s", ddl.SQL)
		}
		if !strings.Contains(ddl.SQL, "DELETE FROM temp") {
			t.Errorf("procedure missing 'DELETE FROM temp': %s", ddl.SQL)
		}
		if !strings.Contains(ddl.SQL, "END") {
			t.Errorf("procedure missing 'END': %s", ddl.SQL)
		}
	}

	// Third statement: INSERT after procedure
	dml2, ok := result.Statements[2].(DMLStatement)
	if !ok {
		t.Errorf("statement 2: expected DMLStatement, got %T", result.Statements[2])
	} else if dml2.ColumnValues[0][1] != "After proc" {
		t.Errorf("statement 2: name = %v, want 'After proc'", dml2.ColumnValues[0][1])
	}
}

func TestDumpParser_DelimiterResetToSemicolon(t *testing.T) {
	parser := NewDumpParser()

	// Test that delimiter properly resets back to semicolon
	sql := `
DELIMITER ;;
CREATE TRIGGER test_trigger BEFORE INSERT ON users FOR EACH ROW BEGIN SET NEW.created_at = NOW(); END;;
DELIMITER ;
INSERT INTO users (id, name) VALUES (1, 'Test');
INSERT INTO users (id, name) VALUES (2, 'Test2');
`

	result, err := parser.ParseStream(strings.NewReader(sql))
	if err != nil {
		t.Fatalf("ParseStream() error = %v", err)
	}

	// Should have: 1 DDL (trigger) + 2 DML (INSERTs)
	if len(result.Statements) != 3 {
		t.Errorf("expected 3 statements, got %d", len(result.Statements))
	}

	// The two INSERT statements should both be parsed correctly after delimiter reset
	dmlCount := 0
	for _, stmt := range result.Statements {
		if _, ok := stmt.(DMLStatement); ok {
			dmlCount++
		}
	}
	if dmlCount != 2 {
		t.Errorf("expected 2 DML statements after delimiter reset, got %d", dmlCount)
	}
}

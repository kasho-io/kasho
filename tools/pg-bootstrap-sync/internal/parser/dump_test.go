package parser

import (
	"strings"
	"testing"
)

func TestDumpParser_ParseStream(t *testing.T) {
	parser := NewDumpParser()
	
	// Test data with DDL and DML statements
	dumpData := `-- Test PostgreSQL dump
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE
);

CREATE INDEX idx_users_email ON users(email);

COPY users (id, name, email) FROM stdin;
1	John Doe	john@example.com
2	Jane Smith	jane@example.com
\.

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200)
);

COPY posts (id, title) FROM stdin;
1	First Post
2	Second Post
\.`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)
	
	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}
	
	// Verify basic counts (3 DDL: CREATE TABLE users, CREATE INDEX, CREATE TABLE posts)
	if len(result.Statements) != 5 { // 3 DDL + 2 DML
		t.Errorf("Expected 5 statements, got %d", len(result.Statements))
	}
	
	if result.Metadata.DDLCount != 3 {
		t.Errorf("Expected 3 DDL statements, got %d", result.Metadata.DDLCount)
	}
	
	if result.Metadata.DMLCount != 2 {
		t.Errorf("Expected 2 DML statements, got %d", result.Metadata.DMLCount)
	}
	
	// Verify tables found
	expectedTables := []string{"users", "posts"}
	if len(result.Metadata.TablesFound) != len(expectedTables) {
		t.Errorf("Expected %d tables, got %d", len(expectedTables), len(result.Metadata.TablesFound))
	}
	
	// Check specific statements
	for _, stmt := range result.Statements {
		switch s := stmt.(type) {
		case DDLStatement:
			if !strings.Contains(s.SQL, "CREATE") {
				t.Errorf("DDL statement should contain CREATE: %s", s.SQL)
			}
		case DMLStatement:
			if s.Table == "users" {
				if len(s.ColumnNames) != 3 {
					t.Errorf("Users table should have 3 columns, got %d", len(s.ColumnNames))
				}
				if len(s.ColumnValues) != 2 {
					t.Errorf("Users table should have 2 rows, got %d", len(s.ColumnValues))
				}
			}
		}
	}
}

func TestDumpParser_ParseCopyStatement(t *testing.T) {
	parser := NewDumpParser()
	
	tests := []struct {
		line     string
		expected *copyInfo
	}{
		{
			line: "COPY users (id, name, email) FROM stdin;",
			expected: &copyInfo{
				table:   "users",
				columns: []string{"id", "name", "email"},
			},
		},
		{
			line: "COPY posts (id, title) FROM stdin;",
			expected: &copyInfo{
				table:   "posts", 
				columns: []string{"id", "title"},
			},
		},
		{
			line: "SELECT * FROM users;",
			expected: nil,
		},
	}
	
	for _, tt := range tests {
		result := parser.parseCopyStatement(tt.line)
		
		if tt.expected == nil {
			if result != nil {
				t.Errorf("Expected nil for line %q, got %+v", tt.line, result)
			}
			continue
		}
		
		if result == nil {
			t.Errorf("Expected result for line %q, got nil", tt.line)
			continue
		}
		
		if result.table != tt.expected.table {
			t.Errorf("Expected table %q, got %q", tt.expected.table, result.table)
		}
		
		if len(result.columns) != len(tt.expected.columns) {
			t.Errorf("Expected %d columns, got %d", len(tt.expected.columns), len(result.columns))
		}
	}
}

func TestDumpParser_UnescapeCopyValue(t *testing.T) {
	parser := NewDumpParser()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"\\N", ""},           // NULL value
		{"hello", "hello"},    // Regular string
		{"hello\\tworld", "hello\tworld"}, // Tab escape
		{"line1\\nline2", "line1\nline2"}, // Newline escape
		{"back\\\\slash", "back\\slash"},  // Backslash escape
	}
	
	for _, tt := range tests {
		result := parser.unescapeCopyValue(tt.input)
		if result != tt.expected {
			t.Errorf("unescapeCopyValue(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDumpParser_ParseValuesList(t *testing.T) {
	parser := NewDumpParser()
	
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "1, 'John Doe', 'john@example.com'",
			expected: []string{"1", "John Doe", "john@example.com"},
		},
		{
			input:    "2, 'Jane''s Post', NULL",
			expected: []string{"2", "Jane's Post", ""},
		},
		{
			input:    "3, 'Text with, comma', true",
			expected: []string{"3", "Text with, comma", "true"},
		},
		{
			input:    "NOW(), 'function call'",
			expected: []string{"NOW()", "function call"},
		},
	}
	
	for _, tt := range tests {
		result := parser.parseValuesList(tt.input)
		
		if len(result) != len(tt.expected) {
			t.Errorf("parseValuesList(%q) length = %d, expected %d", tt.input, len(result), len(tt.expected))
			continue
		}
		
		for i, val := range result {
			if val != tt.expected[i] {
				t.Errorf("parseValuesList(%q) value[%d] = %q, expected %q", tt.input, i, val, tt.expected[i])
			}
		}
	}
}

func TestDumpParser_CleanValue(t *testing.T) {
	parser := NewDumpParser()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"'John Doe'", "John Doe"},
		{"\"quoted string\"", "quoted string"},
		{"NULL", ""},
		{"null", ""},
		{"123", "123"},
		{"'Jane''s Post'", "Jane's Post"},
		{"  'trimmed'  ", "trimmed"},
	}
	
	for _, tt := range tests {
		result := parser.cleanValue(tt.input)
		if result != tt.expected {
			t.Errorf("cleanValue(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDumpParser_ParseStream_WithInserts(t *testing.T) {
	parser := NewDumpParser()
	
	// Test data with DDL and INSERT statements
	dumpData := `-- Test PostgreSQL dump with INSERTs
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE
);

INSERT INTO users (id, name, email) VALUES (1, 'John Doe', 'john@example.com');
INSERT INTO users (id, name, email) VALUES (2, 'Jane Smith', 'jane@example.com');

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200)
);

INSERT INTO posts VALUES (1, 'First Post');
INSERT INTO posts VALUES (2, 'Second Post');`
	
	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)
	
	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}
	
	// Verify basic counts (2 DDL + 4 DML = 6 total statements)
	if len(result.Statements) != 6 {
		t.Errorf("Expected 6 statements, got %d", len(result.Statements))
	}
	
	if result.Metadata.DDLCount != 2 {
		t.Errorf("Expected 2 DDL statements, got %d", result.Metadata.DDLCount)
	}
	
	if result.Metadata.DMLCount != 4 {
		t.Errorf("Expected 4 DML statements, got %d", result.Metadata.DMLCount)
	}
	
	// Verify tables found
	expectedTables := []string{"users", "posts"}
	if len(result.Metadata.TablesFound) != len(expectedTables) {
		t.Errorf("Expected %d tables, got %d", len(expectedTables), len(result.Metadata.TablesFound))
	}
	
	// Count statement types
	ddlCount := 0
	dmlCount := 0
	
	for _, stmt := range result.Statements {
		switch s := stmt.(type) {
		case DDLStatement:
			ddlCount++
			if !strings.Contains(s.SQL, "CREATE") {
				t.Errorf("DDL statement should contain CREATE: %s", s.SQL)
			}
		case DMLStatement:
			dmlCount++
			if s.Table != "users" && s.Table != "posts" {
				t.Errorf("DML statement should be for users or posts table, got %s", s.Table)
			}
			if len(s.ColumnValues) != 1 {
				t.Errorf("DML statement should have 1 row, got %d", len(s.ColumnValues))
			}
		}
	}
	
	if ddlCount != 2 {
		t.Errorf("Expected 2 DDL statements in result, got %d", ddlCount)
	}
	
	if dmlCount != 4 {
		t.Errorf("Expected 4 DML statements in result, got %d", dmlCount)
	}
}
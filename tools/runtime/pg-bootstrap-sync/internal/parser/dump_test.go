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
			line:     "SELECT * FROM users;",
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
		{"\\N", ""},                       // NULL value
		{"hello", "hello"},                // Regular string
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

func TestDumpParser_ComprehensiveStatementTypes(t *testing.T) {
	parser := NewDumpParser()

	// Test comprehensive pg_dump output based on actual dump file
	dumpData := `-- PostgreSQL database dump
SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- DDL: CREATE FUNCTION
CREATE FUNCTION public.test_capture_ddl_command() RETURNS event_trigger
    LANGUAGE plpgsql
    AS $$
  BEGIN
    PERFORM set_config('ddl.command', current_query(), true);
  END;
  $$;

-- DDL: CREATE TABLE
CREATE TABLE public.test_ddl_log (
    id integer NOT NULL,
    lsn pg_lsn,
    ts timestamp with time zone DEFAULT now(),
    username text,
    database text,
    ddl text
);

-- DDL: CREATE SEQUENCE
CREATE SEQUENCE public.test_ddl_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1;

-- DDL: ALTER SEQUENCE
ALTER SEQUENCE public.test_ddl_log_id_seq OWNED BY public.test_ddl_log.id;

-- DDL: ALTER TABLE
ALTER TABLE ONLY public.test_ddl_log ALTER COLUMN id SET DEFAULT nextval('public.test_ddl_log_id_seq'::regclass);

-- DDL: CREATE EVENT TRIGGER
CREATE EVENT TRIGGER test_capture_ddl ON ddl_command_start
   EXECUTE FUNCTION public.test_capture_ddl_command();

-- DDL: CREATE TRIGGER
CREATE TRIGGER test_cleanup_ddl_logs_trigger AFTER INSERT ON public.test_ddl_log 
   FOR EACH ROW EXECUTE FUNCTION public.test_trigger_cleanup_ddl_logs();

-- DDL: COMMENT
COMMENT ON SCHEMA public IS 'standard public schema';

-- These should be skipped
CREATE PUBLICATION kasho_pub FOR ALL TABLES;
DROP PUBLICATION IF EXISTS old_pub;

-- DML: INSERT
INSERT INTO public.test_ddl_log (id, lsn) VALUES (1, '0/1A34588');

-- DDL: SELECT setval
SELECT pg_catalog.setval('public.test_ddl_log_id_seq', 1, true);`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Count different statement types
	statementTypes := make(map[string]int)
	for _, stmt := range result.Statements {
		switch s := stmt.(type) {
		case DDLStatement:
			// Try to identify the specific type from SQL
			sql := strings.ToUpper(strings.TrimSpace(s.SQL))
			if strings.HasPrefix(sql, "CREATE FUNCTION") {
				statementTypes["CREATE_FUNCTION"]++
			} else if strings.HasPrefix(sql, "CREATE TABLE") {
				statementTypes["CREATE_TABLE"]++
			} else if strings.HasPrefix(sql, "CREATE SEQUENCE") {
				statementTypes["CREATE_SEQUENCE"]++
			} else if strings.HasPrefix(sql, "ALTER SEQUENCE") {
				statementTypes["ALTER_SEQUENCE"]++
			} else if strings.HasPrefix(sql, "ALTER TABLE") {
				statementTypes["ALTER_TABLE"]++
			} else if strings.HasPrefix(sql, "CREATE EVENT TRIGGER") {
				statementTypes["CREATE_EVENT_TRIGGER"]++
			} else if strings.HasPrefix(sql, "CREATE TRIGGER") {
				statementTypes["CREATE_TRIGGER"]++
			} else if strings.HasPrefix(sql, "COMMENT") {
				statementTypes["COMMENT"]++
			} else if strings.Contains(sql, "SETVAL") {
				statementTypes["SELECT_SETVAL"]++
			} else {
				statementTypes["OTHER_DDL"]++
			}
		case DMLStatement:
			statementTypes["DML"]++
		}
	}

	// Verify expected counts
	expectedCounts := map[string]int{
		"CREATE_FUNCTION":      1,
		"CREATE_TABLE":         1,
		"CREATE_SEQUENCE":      1,
		"ALTER_SEQUENCE":       1,
		"ALTER_TABLE":          1,
		"CREATE_EVENT_TRIGGER": 1,
		"CREATE_TRIGGER":       1,
		"COMMENT":              1,
		"SELECT_SETVAL":        1,
		"DML":                  1,
	}

	for stmtType, expectedCount := range expectedCounts {
		if statementTypes[stmtType] != expectedCount {
			t.Errorf("Expected %d %s statements, got %d", expectedCount, stmtType, statementTypes[stmtType])
		}
	}

	// Verify publication statements were skipped
	if statementTypes["CREATE_PUBLICATION"] > 0 {
		t.Error("CREATE PUBLICATION should have been skipped")
	}
	if statementTypes["DROP_PUBLICATION"] > 0 {
		t.Error("DROP PUBLICATION should have been skipped")
	}

	// Verify total counts
	if result.Metadata.DDLCount != 9 { // All DDL except publications
		t.Errorf("Expected 9 DDL statements, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 1 {
		t.Errorf("Expected 1 DML statement, got %d", result.Metadata.DMLCount)
	}
}

func TestDumpParser_AllStatementTypes(t *testing.T) {
	parser := NewDumpParser()

	// Test data with all supported statement types
	dumpData := `-- Comprehensive test with all pg_dump statement types
-- DDL: CREATE TABLE
CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT);

-- DDL: CREATE SEQUENCE
CREATE SEQUENCE user_id_seq START 1000;

-- DDL: ALTER SEQUENCE
ALTER SEQUENCE user_id_seq OWNED BY users.id;

-- DDL: CREATE INDEX
CREATE INDEX idx_users_username ON users(username);

-- DDL: ALTER TABLE
ALTER TABLE users ADD COLUMN email TEXT;

-- DDL: COMMENT
COMMENT ON TABLE users IS 'User accounts table';

-- DDL: GRANT
GRANT SELECT ON users TO readonly_user;

-- DML: INSERT
INSERT INTO users (id, username) VALUES (1, 'alice');
INSERT INTO users (id, username) VALUES (2, 'bob');

-- DDL: SELECT with setval (special case for sequences)
SELECT pg_catalog.setval('user_id_seq', 1002, true);

-- Session control (should be skipped)
SET statement_timeout = 0;
SET lock_timeout = 0;

-- COPY data (DML)
COPY users (id, username, email) FROM stdin;
3	charlie	charlie@example.com
4	david	david@example.com
\.

-- DDL: CREATE SEQUENCE for testing  
CREATE SEQUENCE order_id_seq;

-- DDL: TRUNCATE
TRUNCATE TABLE old_data;

-- DDL: DROP
DROP TABLE IF EXISTS temp_table;`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Count statement types
	ddlCount := 0
	dmlCount := 0
	var statementTypes []string

	for _, stmt := range result.Statements {
		switch s := stmt.(type) {
		case DDLStatement:
			ddlCount++
			// Extract statement type from SQL
			sql := strings.ToUpper(strings.TrimSpace(s.SQL))
			if strings.HasPrefix(sql, "CREATE TABLE") {
				statementTypes = append(statementTypes, "CREATE_TABLE")
			} else if strings.HasPrefix(sql, "CREATE SEQUENCE") {
				statementTypes = append(statementTypes, "CREATE_SEQUENCE")
			} else if strings.HasPrefix(sql, "ALTER SEQUENCE") {
				statementTypes = append(statementTypes, "ALTER_SEQUENCE")
			} else if strings.HasPrefix(sql, "CREATE INDEX") {
				statementTypes = append(statementTypes, "CREATE_INDEX")
			} else if strings.HasPrefix(sql, "ALTER TABLE") {
				statementTypes = append(statementTypes, "ALTER_TABLE")
			} else if strings.HasPrefix(sql, "COMMENT") {
				statementTypes = append(statementTypes, "COMMENT")
			} else if strings.HasPrefix(sql, "GRANT") {
				statementTypes = append(statementTypes, "GRANT")
			} else if strings.Contains(sql, "SETVAL") {
				statementTypes = append(statementTypes, "SELECT_SETVAL")
			} else if strings.HasPrefix(sql, "TRUNCATE") {
				statementTypes = append(statementTypes, "TRUNCATE")
			} else if strings.HasPrefix(sql, "DROP") {
				statementTypes = append(statementTypes, "DROP")
			}
		case DMLStatement:
			dmlCount++
			if len(s.ColumnValues) > 0 {
				statementTypes = append(statementTypes, "DML")
			}
		}
	}

	// Verify counts
	if result.Metadata.DDLCount != 11 { // All DDL statements including setval
		t.Errorf("Expected 11 DDL statements, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 3 { // 2 INSERTs + 1 COPY
		t.Errorf("Expected 3 DML statements, got %d", result.Metadata.DMLCount)
	}

	// Verify all expected statement types were parsed
	expectedTypes := map[string]bool{
		"CREATE_TABLE":    false,
		"CREATE_SEQUENCE": false,
		"ALTER_SEQUENCE":  false,
		"CREATE_INDEX":    false,
		"ALTER_TABLE":     false,
		"COMMENT":         false,
		"GRANT":           false,
		"SELECT_SETVAL":   false,
		"TRUNCATE":        false,
		"DROP":            false,
		"DML":             false,
	}

	for _, st := range statementTypes {
		expectedTypes[st] = true
	}

	for stype, found := range expectedTypes {
		if !found {
			t.Errorf("Statement type %s was not parsed", stype)
		}
	}
}

func TestDumpParser_UnsupportedStatements(t *testing.T) {
	parser := NewDumpParser()

	testCases := []struct {
		name        string
		dumpData    string
		expectedErr string
	}{
		{
			name: "UPDATE statement",
			dumpData: `CREATE TABLE test (id INT);
UPDATE test SET id = 2 WHERE id = 1;`,
			expectedErr: "unexpected DML statement type UPDATE in pg_dump",
		},
		{
			name: "DELETE statement",
			dumpData: `CREATE TABLE test (id INT);
DELETE FROM test WHERE id = 1;`,
			expectedErr: "unexpected DML statement type DELETE in pg_dump",
		},
		{
			name: "Non-setval SELECT",
			dumpData: `CREATE TABLE test (id INT);
SELECT * FROM test WHERE id = 1;`,
			expectedErr: "unexpected SELECT statement in pg_dump (not a setval)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.dumpData)
			_, err := parser.ParseStream(reader)

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
				return
			}

			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Expected error containing %q, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestDumpParser_SessionControlStatements(t *testing.T) {
	parser := NewDumpParser()

	// Test that SET and transaction control statements are skipped
	dumpData := `-- Test with session control statements
SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SELECT pg_catalog.set_config('search_path', '', false);
BEGIN;

CREATE TABLE test (id INT);
INSERT INTO test VALUES (1);

COMMIT;`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Should only have CREATE TABLE and INSERT, SET/BEGIN/COMMIT/set_config should be skipped
	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements (CREATE and INSERT), got %d", len(result.Statements))
	}

	if result.Metadata.DDLCount != 1 {
		t.Errorf("Expected 1 DDL statement, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 1 {
		t.Errorf("Expected 1 DML statement, got %d", result.Metadata.DMLCount)
	}
}

func TestDumpParser_SetvalSpecialCase(t *testing.T) {
	parser := NewDumpParser()

	// Test that setval SELECT statements are treated as DDL
	dumpData := `-- Test setval handling
CREATE SEQUENCE test_seq;
SELECT pg_catalog.setval('test_seq', 100);
SELECT pg_catalog.setval('test_seq', 200, true);
SELECT pg_catalog.setval('test_seq', 300, false);`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// All statements should be DDL (CREATE SEQUENCE + 3 setval calls)
	if result.Metadata.DDLCount != 4 {
		t.Errorf("Expected 4 DDL statements, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 0 {
		t.Errorf("Expected 0 DML statements, got %d", result.Metadata.DMLCount)
	}

	// Verify all statements are DDL
	for i, stmt := range result.Statements {
		if _, ok := stmt.(DDLStatement); !ok {
			t.Errorf("Statement %d should be DDL, got %T", i, stmt)
		}
	}
}

func TestDumpParser_SelectStatements(t *testing.T) {
	parser := NewDumpParser()

	testCases := []struct {
		name        string
		dumpData    string
		expectError bool
		expectedDDL int
	}{
		{
			name: "setval statements",
			dumpData: `CREATE SEQUENCE test_seq;
SELECT pg_catalog.setval('test_seq', 100);`,
			expectError: false,
			expectedDDL: 2,
		},
		{
			name: "set_config statement",
			dumpData: `SELECT pg_catalog.set_config('search_path', '', false);
CREATE TABLE test (id INT);`,
			expectError: false,
			expectedDDL: 1, // Only CREATE TABLE, set_config is skipped
		},
		{
			name: "mixed set_config and setval",
			dumpData: `SELECT pg_catalog.set_config('search_path', '', false);
CREATE SEQUENCE seq;
SELECT pg_catalog.setval('seq', 1);`,
			expectError: false,
			expectedDDL: 2, // CREATE SEQUENCE and setval
		},
		{
			name: "invalid SELECT",
			dumpData: `CREATE TABLE test (id INT);
SELECT * FROM test;`,
			expectError: true,
			expectedDDL: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.dumpData)
			result, err := parser.ParseStream(reader)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result.Metadata.DDLCount != tc.expectedDDL {
					t.Errorf("Expected %d DDL statements, got %d", tc.expectedDDL, result.Metadata.DDLCount)
				}
			}
		})
	}
}

func TestDumpParser_PublicationStatements(t *testing.T) {
	parser := NewDumpParser()

	// Test that publication/subscription statements are skipped
	dumpData := `-- Test with publication/subscription statements
CREATE TABLE test (id INT);

-- These should all be skipped
CREATE PUBLICATION kasho_pub FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');
ALTER PUBLICATION kasho_pub SET (publish = 'insert, update');
DROP PUBLICATION IF EXISTS old_pub;

CREATE SUBSCRIPTION kasho_sub CONNECTION 'host=primary' PUBLICATION kasho_pub;
ALTER SUBSCRIPTION kasho_sub DISABLE;
DROP SUBSCRIPTION IF EXISTS old_sub;

-- This should be processed
INSERT INTO test VALUES (1);`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Should only have CREATE TABLE and INSERT
	if len(result.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(result.Statements))
	}

	if result.Metadata.DDLCount != 1 {
		t.Errorf("Expected 1 DDL statement, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 1 {
		t.Errorf("Expected 1 DML statement, got %d", result.Metadata.DMLCount)
	}
}

func TestDumpParser_EventTriggers(t *testing.T) {
	parser := NewDumpParser()

	// Test event trigger statements
	dumpData := `-- Test event triggers
CREATE TABLE test (id INT);

-- Function for event trigger
CREATE FUNCTION capture_ddl() RETURNS event_trigger AS $$
BEGIN
    RAISE NOTICE 'DDL command: %', tg_tag;
END;
$$ LANGUAGE plpgsql;

-- Event trigger
CREATE EVENT TRIGGER my_capture_ddl ON ddl_command_start
   EXECUTE FUNCTION capture_ddl();

-- Regular trigger for comparison
CREATE TRIGGER test_trigger AFTER INSERT ON test
   FOR EACH ROW EXECUTE FUNCTION capture_ddl();

INSERT INTO test VALUES (1);`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Should have CREATE TABLE, CREATE FUNCTION, CREATE EVENT TRIGGER, CREATE TRIGGER, and INSERT
	if len(result.Statements) != 5 {
		t.Errorf("Expected 5 statements, got %d", len(result.Statements))
	}

	if result.Metadata.DDLCount != 4 {
		t.Errorf("Expected 4 DDL statements, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 1 {
		t.Errorf("Expected 1 DML statement, got %d", result.Metadata.DMLCount)
	}

	// Verify we have both trigger types
	eventTriggerFound := false
	regularTriggerFound := false
	for _, stmt := range result.Statements {
		if ddl, ok := stmt.(DDLStatement); ok {
			if strings.Contains(ddl.SQL, "CREATE EVENT TRIGGER") {
				eventTriggerFound = true
			} else if strings.Contains(ddl.SQL, "CREATE TRIGGER") && !strings.Contains(ddl.SQL, "EVENT") {
				regularTriggerFound = true
			}
		}
	}

	if !eventTriggerFound {
		t.Error("Event trigger not found in statements")
	}
	if !regularTriggerFound {
		t.Error("Regular trigger not found in statements")
	}
}

func TestDumpParser_InsertWithoutColumns(t *testing.T) {
	parser := NewDumpParser()

	// Test INSERT statements without explicit column names
	dumpData := `-- Test INSERT without column names
CREATE TABLE test (id INT, name TEXT, email TEXT);

-- This INSERT has column names (good)
INSERT INTO test (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- This INSERT lacks column names (problematic)
INSERT INTO test VALUES (2, 'Bob', 'bob@example.com');`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Should have 1 DDL and 2 DML
	if len(result.Statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(result.Statements))
	}

	// Check the INSERT statements
	dmlCount := 0
	for _, stmt := range result.Statements {
		if dml, ok := stmt.(DMLStatement); ok {
			dmlCount++
			if dml.Table != "test" {
				t.Errorf("Expected table 'test', got %s", dml.Table)
			}

			// First INSERT should have column names
			if dmlCount == 1 && len(dml.ColumnNames) != 3 {
				t.Errorf("First INSERT should have 3 column names, got %d", len(dml.ColumnNames))
			}

			// Second INSERT will have empty column names (this is the issue)
			if dmlCount == 2 && len(dml.ColumnNames) != 0 {
				t.Errorf("Second INSERT should have 0 column names, got %d", len(dml.ColumnNames))
			}

			// Both should have 3 values
			if len(dml.ColumnValues) != 1 || len(dml.ColumnValues[0]) != 3 {
				t.Errorf("Expected 1 row with 3 values, got %d rows", len(dml.ColumnValues))
			}
		}
	}

	if dmlCount != 2 {
		t.Errorf("Expected 2 DML statements, got %d", dmlCount)
	}
}

func TestDumpParser_DollarQuotedStrings(t *testing.T) {
	parser := NewDumpParser()

	// Test multi-line statements with dollar-quoted strings
	dumpData := `-- Test dollar-quoted strings
CREATE TABLE test (id INT);

-- Function with dollar quotes
CREATE FUNCTION update_modified_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Another function with named dollar quotes
CREATE OR REPLACE FUNCTION complex_function(input text) RETURNS text AS $func$
DECLARE
    result text;
BEGIN
    -- This semicolon should not end the statement;
    result := input || '; processed';
    RETURN result;
END;
$func$ LANGUAGE plpgsql;

-- Regular statement after functions
CREATE INDEX idx_test ON test(id);

-- Function with nested dollar quotes
CREATE FUNCTION nested_quotes() RETURNS text AS $$
BEGIN
    RETURN $inner$This is a string with ; semicolon$inner$;
END;
$$ LANGUAGE plpgsql;`

	reader := strings.NewReader(dumpData)
	result, err := parser.ParseStream(reader)

	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	// Should have 5 DDL statements (CREATE TABLE, 3 functions, CREATE INDEX)
	if result.Metadata.DDLCount != 5 {
		t.Errorf("Expected 5 DDL statements, got %d", result.Metadata.DDLCount)
	}

	if result.Metadata.DMLCount != 0 {
		t.Errorf("Expected 0 DML statements, got %d", result.Metadata.DMLCount)
	}

	// Verify the function statements were parsed correctly
	functionCount := 0
	for _, stmt := range result.Statements {
		if ddl, ok := stmt.(DDLStatement); ok {
			if strings.Contains(strings.ToUpper(ddl.SQL), "CREATE FUNCTION") ||
				strings.Contains(strings.ToUpper(ddl.SQL), "CREATE OR REPLACE FUNCTION") {
				functionCount++
				// Verify the full function body is included
				if !strings.Contains(ddl.SQL, "BEGIN") || !strings.Contains(ddl.SQL, "END;") {
					t.Errorf("Function statement appears truncated: %s", ddl.SQL)
				}
			}
		}
	}

	if functionCount != 3 {
		t.Errorf("Expected 3 function statements, found %d", functionCount)
	}
}

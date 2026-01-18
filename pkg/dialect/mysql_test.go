package dialect

import (
	"testing"
	"time"

	"kasho/proto"
)

func TestMySQL_Name(t *testing.T) {
	d := NewMySQL()
	if got := d.Name(); got != "mysql" {
		t.Errorf("Name() = %v, want mysql", got)
	}
}

func TestMySQL_GetDriverName(t *testing.T) {
	d := NewMySQL()
	if got := d.GetDriverName(); got != "mysql" {
		t.Errorf("GetDriverName() = %v, want mysql", got)
	}
}

func TestMySQL_QuoteIdentifier(t *testing.T) {
	d := NewMySQL()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple name", "users", "`users`"},
		{"with space", "user table", "`user table`"},
		{"with backtick", "user`name", "`user``name`"},
		{"reserved word", "select", "`select`"},
		{"mixed case", "UserTable", "`UserTable`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.QuoteIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("QuoteIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMySQL_FormatValue(t *testing.T) {
	d := NewMySQL()

	tests := []struct {
		name    string
		value   *proto.ColumnValue
		want    string
		wantErr bool
	}{
		{
			name:  "nil value",
			value: nil,
			want:  "NULL",
		},
		{
			name:  "nil inner value",
			value: &proto.ColumnValue{Value: nil},
			want:  "NULL",
		},
		{
			name:  "string value",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "hello"}},
			want:  "'hello'",
		},
		{
			name:  "string with single quote",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "it's"}},
			want:  "'it''s'",
		},
		{
			name:  "string with backslash",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "path\\to\\file"}},
			want:  "'path\\\\to\\\\file'",
		},
		{
			name:  "string with quote and backslash",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "it's a \\test"}},
			want:  "'it''s a \\\\test'",
		},
		{
			name:  "integer value",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 42}},
			want:  "42",
		},
		{
			name:  "negative integer",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: -100}},
			want:  "-100",
		},
		{
			name:  "float value",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: 3.14}},
			want:  "3.140000",
		},
		{
			name:  "bool true (MySQL uses 1)",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
			want:  "1",
		},
		{
			name:  "bool false (MySQL uses 0)",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: false}},
			want:  "0",
		},
		{
			name:  "timestamp RFC3339",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"}},
			want:  "'2024-03-20 15:04:05'",
		},
		{
			name:  "date only",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20"}},
			want:  "'2024-03-20'",
		},
		{
			name:    "invalid timestamp",
			value:   &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "not-a-date"}},
			wantErr: true,
		},
		// Edge cases for strings
		{
			name:  "string with newline",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "line1\nline2"}},
			want:  "'line1\nline2'",
		},
		{
			name:  "string with tab",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "col1\tcol2"}},
			want:  "'col1\tcol2'",
		},
		{
			name:  "unicode string",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "日本語テスト"}},
			want:  "'日本語テスト'",
		},
		{
			name:  "emoji string",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "hello 👋 world"}},
			want:  "'hello 👋 world'",
		},
		{
			name:  "empty string",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: ""}},
			want:  "''",
		},
		{
			name:  "string of only quotes",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "'''"}},
			want:  "''''''''",
		},
		{
			name:  "string with mixed escapes",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "it's a \\path\\with 'quotes'"}},
			want:  "'it''s a \\\\path\\\\with ''quotes'''",
		},
		// Edge cases for integers
		{
			name:  "max int64",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 9223372036854775807}},
			want:  "9223372036854775807",
		},
		{
			name:  "min int64",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: -9223372036854775808}},
			want:  "-9223372036854775808",
		},
		{
			name:  "zero",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 0}},
			want:  "0",
		},
		// Edge cases for floats
		{
			name:  "very small float",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: 0.000001}},
			want:  "0.000001",
		},
		{
			name:  "negative float",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: -3.14}},
			want:  "-3.140000",
		},
		{
			name:  "zero float",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: 0.0}},
			want:  "0.000000",
		},
		// Edge cases for timestamps
		// Note: timestamps with timezone are preserved as-is (not converted to UTC)
		{
			name:  "timestamp with positive timezone",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05+02:00"}},
			want:  "'2024-03-20 15:04:05'",
		},
		{
			name:  "timestamp with negative timezone",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05-05:00"}},
			want:  "'2024-03-20 15:04:05'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.FormatValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FormatValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMySQL_GetUserTablesQuery(t *testing.T) {
	d := NewMySQL()
	query := d.GetUserTablesQuery()

	if query == "" {
		t.Error("GetUserTablesQuery() returned empty string")
	}

	// Should exclude system schemas
	if !contains(query, "mysql") {
		t.Error("GetUserTablesQuery() should exclude mysql schema")
	}
	if !contains(query, "information_schema") {
		t.Error("GetUserTablesQuery() should exclude information_schema")
	}
	if !contains(query, "performance_schema") {
		t.Error("GetUserTablesQuery() should exclude performance_schema")
	}
}

func TestMySQL_BooleanDiffersFromPostgres(t *testing.T) {
	mysql := NewMySQL()
	postgres := NewPostgreSQL()

	trueVal := &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: true}}
	falseVal := &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: false}}

	mysqlTrue, _ := mysql.FormatValue(trueVal)
	mysqlFalse, _ := mysql.FormatValue(falseVal)
	pgTrue, _ := postgres.FormatValue(trueVal)
	pgFalse, _ := postgres.FormatValue(falseVal)

	// MySQL uses 1/0, PostgreSQL uses true/false
	if mysqlTrue != "1" {
		t.Errorf("MySQL true = %v, want 1", mysqlTrue)
	}
	if mysqlFalse != "0" {
		t.Errorf("MySQL false = %v, want 0", mysqlFalse)
	}
	if pgTrue != "true" {
		t.Errorf("PostgreSQL true = %v, want true", pgTrue)
	}
	if pgFalse != "false" {
		t.Errorf("PostgreSQL false = %v, want false", pgFalse)
	}
}

func TestMySQL_QuoteIdentifierDiffersFromPostgres(t *testing.T) {
	mysql := NewMySQL()
	postgres := NewPostgreSQL()

	// MySQL uses backticks, PostgreSQL uses double quotes
	mysqlQuoted := mysql.QuoteIdentifier("table")
	pgQuoted := postgres.QuoteIdentifier("table")

	if mysqlQuoted != "`table`" {
		t.Errorf("MySQL QuoteIdentifier = %v, want `table`", mysqlQuoted)
	}
	if pgQuoted != `"table"` {
		t.Errorf("PostgreSQL QuoteIdentifier = %v, want \"table\"", pgQuoted)
	}
}

func TestMySQL_FormatDSN(t *testing.T) {
	d := NewMySQL()

	tests := []struct {
		name     string
		connStr  string
		expected string
	}{
		{
			name:     "full URL with port",
			connStr:  "mysql://user:password@localhost:3306/mydb",
			expected: "user:password@tcp(localhost:3306)/mydb",
		},
		{
			name:     "URL without port uses default",
			connStr:  "mysql://user:password@localhost/mydb",
			expected: "user:password@tcp(localhost:3306)/mydb",
		},
		{
			name:     "URL with query parameters",
			connStr:  "mysql://user:password@localhost:3306/mydb?parseTime=true&charset=utf8mb4",
			expected: "user:password@tcp(localhost:3306)/mydb?parseTime=true&charset=utf8mb4",
		},
		{
			name:     "URL with special characters in password",
			connStr:  "mysql://user:p%40ssword@localhost:3306/mydb",
			expected: "user:p%40ssword@tcp(localhost:3306)/mydb",
		},
		{
			name:     "already in DSN format (no mysql:// prefix)",
			connStr:  "user:password@tcp(localhost:3306)/mydb",
			expected: "user:password@tcp(localhost:3306)/mydb",
		},
		{
			name:     "URL with hostname",
			connStr:  "mysql://root:secret@mysql-replica:3306/saas_demo",
			expected: "root:secret@tcp(mysql-replica:3306)/saas_demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.FormatDSN(tt.connStr)
			if got != tt.expected {
				t.Errorf("FormatDSN(%q) = %q, want %q", tt.connStr, got, tt.expected)
			}
		})
	}
}

// Tests for native type formatting methods

func TestMySQL_FormatString(t *testing.T) {
	d := NewMySQL()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple string", "hello", "'hello'"},
		{"with single quote", "it's", "'it''s'"},
		{"with multiple quotes", "it's a 'test'", "'it''s a ''test'''"},
		{"empty string", "", "''"},
		{"with backslash", "path\\to\\file", "'path\\\\to\\\\file'"}, // MySQL escapes backslashes
		{"with quote and backslash", "it's a \\test", "'it''s a \\\\test'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.FormatString(tt.input)
			if got != tt.want {
				t.Errorf("FormatString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMySQL_FormatInt(t *testing.T) {
	d := NewMySQL()

	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"positive", 42, "42"},
		{"negative", -100, "-100"},
		{"zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.FormatInt(tt.input)
			if got != tt.want {
				t.Errorf("FormatInt(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMySQL_FormatBool(t *testing.T) {
	d := NewMySQL()

	if got := d.FormatBool(true); got != "1" {
		t.Errorf("FormatBool(true) = %v, want 1", got)
	}
	if got := d.FormatBool(false); got != "0" {
		t.Errorf("FormatBool(false) = %v, want 0", got)
	}
}

func TestMySQL_FormatTimestamp(t *testing.T) {
	d := NewMySQL()
	ts := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)

	got := d.FormatTimestamp(ts)
	want := "'2024-03-20 15:04:05'"
	if got != want {
		t.Errorf("FormatTimestamp() = %v, want %v", got, want)
	}
}

func TestMySQL_FormatDate(t *testing.T) {
	d := NewMySQL()
	ts := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)

	got := d.FormatDate(ts)
	want := "'2024-03-20'"
	if got != want {
		t.Errorf("FormatDate() = %v, want %v", got, want)
	}
}

func TestMySQL_FormatNull(t *testing.T) {
	d := NewMySQL()
	if got := d.FormatNull(); got != "NULL" {
		t.Errorf("FormatNull() = %v, want NULL", got)
	}
}

// Tests for DDL type methods

func TestMySQL_TypeUUID(t *testing.T) {
	d := NewMySQL()
	if got := d.TypeUUID(); got != "CHAR(36)" {
		t.Errorf("TypeUUID() = %v, want CHAR(36)", got)
	}
}

func TestMySQL_TypeText(t *testing.T) {
	d := NewMySQL()
	if got := d.TypeText(); got != "TEXT" {
		t.Errorf("TypeText() = %v, want TEXT", got)
	}
}

func TestMySQL_TypeTimestamp(t *testing.T) {
	d := NewMySQL()
	if got := d.TypeTimestamp(); got != "DATETIME(6)" {
		t.Errorf("TypeTimestamp() = %v, want DATETIME(6)", got)
	}
}

func TestMySQL_TypeDecimal(t *testing.T) {
	d := NewMySQL()
	if got := d.TypeDecimal(10, 2); got != "DECIMAL(10,2)" {
		t.Errorf("TypeDecimal(10, 2) = %v, want DECIMAL(10,2)", got)
	}
}

func TestMySQL_TypeInteger(t *testing.T) {
	d := NewMySQL()
	if got := d.TypeInteger(); got != "INT" {
		t.Errorf("TypeInteger() = %v, want INT", got)
	}
}

// Test that MySQL and PostgreSQL format methods differ as expected

func TestMySQL_FormatTimestamp_DiffersFromPostgres(t *testing.T) {
	mysql := NewMySQL()
	postgres := NewPostgreSQL()
	ts := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)

	mysqlTs := mysql.FormatTimestamp(ts)
	pgTs := postgres.FormatTimestamp(ts)

	// MySQL uses space-separated format
	if mysqlTs != "'2024-03-20 15:04:05'" {
		t.Errorf("MySQL FormatTimestamp() = %v, want '2024-03-20 15:04:05'", mysqlTs)
	}
	// PostgreSQL uses RFC3339 format
	if pgTs != "'2024-03-20T15:04:05Z'" {
		t.Errorf("PostgreSQL FormatTimestamp() = %v, want '2024-03-20T15:04:05Z'", pgTs)
	}
}

func TestMySQL_TypeUUID_DiffersFromPostgres(t *testing.T) {
	mysql := NewMySQL()
	postgres := NewPostgreSQL()

	if mysql.TypeUUID() != "CHAR(36)" {
		t.Errorf("MySQL TypeUUID() = %v, want CHAR(36)", mysql.TypeUUID())
	}
	if postgres.TypeUUID() != "UUID" {
		t.Errorf("PostgreSQL TypeUUID() = %v, want UUID", postgres.TypeUUID())
	}
}

func TestMySQL_TypeTimestamp_DiffersFromPostgres(t *testing.T) {
	mysql := NewMySQL()
	postgres := NewPostgreSQL()

	if mysql.TypeTimestamp() != "DATETIME(6)" {
		t.Errorf("MySQL TypeTimestamp() = %v, want DATETIME(6)", mysql.TypeTimestamp())
	}
	if postgres.TypeTimestamp() != "TIMESTAMP WITH TIME ZONE" {
		t.Errorf("PostgreSQL TypeTimestamp() = %v, want TIMESTAMP WITH TIME ZONE", postgres.TypeTimestamp())
	}
}

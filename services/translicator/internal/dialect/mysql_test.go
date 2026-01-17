package dialect

import (
	"testing"

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

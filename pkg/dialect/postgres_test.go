package dialect

import (
	"testing"
	"time"

	"kasho/proto"
)

func TestPostgreSQL_Name(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.Name(); got != "postgresql" {
		t.Errorf("Name() = %v, want postgresql", got)
	}
}

func TestPostgreSQL_GetDriverName(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.GetDriverName(); got != "postgres" {
		t.Errorf("GetDriverName() = %v, want postgres", got)
	}
}

func TestPostgreSQL_QuoteIdentifier(t *testing.T) {
	d := NewPostgreSQL()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple name", "users", `"users"`},
		{"with space", "user table", `"user table"`},
		{"with double quote", `user"name`, `"user""name"`},
		{"reserved word", "select", `"select"`},
		{"mixed case", "UserTable", `"UserTable"`},
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

func TestPostgreSQL_FormatValue(t *testing.T) {
	d := NewPostgreSQL()

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
			name:  "string with multiple quotes",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "it's a 'test'"}},
			want:  "'it''s a ''test'''",
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
			name:  "bool true",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
			want:  "true",
		},
		{
			name:  "bool false",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: false}},
			want:  "false",
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
			name:  "string with backslash",
			value: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "path\\to\\file"}},
			want:  "'path\\to\\file'",
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

func TestPostgreSQL_GetUserTablesQuery(t *testing.T) {
	d := NewPostgreSQL()
	query := d.GetUserTablesQuery()

	if query == "" {
		t.Error("GetUserTablesQuery() returned empty string")
	}

	// Should exclude system schemas
	if !contains(query, "pg_catalog") {
		t.Error("GetUserTablesQuery() should exclude pg_catalog")
	}
	if !contains(query, "information_schema") {
		t.Error("GetUserTablesQuery() should exclude information_schema")
	}
}

func TestPostgreSQL_formatRegclass(t *testing.T) {
	d := NewPostgreSQL()

	tests := []struct {
		name    string
		schema  string
		seqName string
		want    string
	}{
		{"simple names", "public", "users_id_seq", "'public.users_id_seq'"},
		{"with single quote in schema", "it's_schema", "seq", "'it''s_schema.seq'"},
		{"with single quote in sequence", "public", "user's_seq", "'public.user''s_seq'"},
		{"with quotes in both", "it's", "also'quoted", "'it''s.also''quoted'"},
		{"reserved word schema", "select", "from_seq", "'select.from_seq'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.formatRegclass(tt.schema, tt.seqName)
			if got != tt.want {
				t.Errorf("formatRegclass(%q, %q) = %v, want %v", tt.schema, tt.seqName, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for native type formatting methods

func TestPostgreSQL_FormatString(t *testing.T) {
	d := NewPostgreSQL()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple string", "hello", "'hello'"},
		{"with single quote", "it's", "'it''s'"},
		{"with multiple quotes", "it's a 'test'", "'it''s a ''test'''"},
		{"empty string", "", "''"},
		{"with backslash", "path\\to\\file", "'path\\to\\file'"}, // PostgreSQL doesn't escape backslashes
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

func TestPostgreSQL_FormatInt(t *testing.T) {
	d := NewPostgreSQL()

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

func TestPostgreSQL_FormatBool(t *testing.T) {
	d := NewPostgreSQL()

	if got := d.FormatBool(true); got != "true" {
		t.Errorf("FormatBool(true) = %v, want true", got)
	}
	if got := d.FormatBool(false); got != "false" {
		t.Errorf("FormatBool(false) = %v, want false", got)
	}
}

func TestPostgreSQL_FormatTimestamp(t *testing.T) {
	d := NewPostgreSQL()
	ts := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)

	got := d.FormatTimestamp(ts)
	want := "'2024-03-20T15:04:05Z'"
	if got != want {
		t.Errorf("FormatTimestamp() = %v, want %v", got, want)
	}
}

func TestPostgreSQL_FormatDate(t *testing.T) {
	d := NewPostgreSQL()
	ts := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)

	got := d.FormatDate(ts)
	want := "'2024-03-20'"
	if got != want {
		t.Errorf("FormatDate() = %v, want %v", got, want)
	}
}

func TestPostgreSQL_FormatNull(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.FormatNull(); got != "NULL" {
		t.Errorf("FormatNull() = %v, want NULL", got)
	}
}

// Tests for DDL type methods

func TestPostgreSQL_TypeUUID(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.TypeUUID(); got != "UUID" {
		t.Errorf("TypeUUID() = %v, want UUID", got)
	}
}

func TestPostgreSQL_TypeText(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.TypeText(); got != "TEXT" {
		t.Errorf("TypeText() = %v, want TEXT", got)
	}
}

func TestPostgreSQL_TypeTimestamp(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.TypeTimestamp(); got != "TIMESTAMP WITH TIME ZONE" {
		t.Errorf("TypeTimestamp() = %v, want TIMESTAMP WITH TIME ZONE", got)
	}
}

func TestPostgreSQL_TypeDecimal(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.TypeDecimal(10, 2); got != "DECIMAL(10,2)" {
		t.Errorf("TypeDecimal(10, 2) = %v, want DECIMAL(10,2)", got)
	}
}

func TestPostgreSQL_TypeInteger(t *testing.T) {
	d := NewPostgreSQL()
	if got := d.TypeInteger(); got != "INTEGER" {
		t.Errorf("TypeInteger() = %v, want INTEGER", got)
	}
}

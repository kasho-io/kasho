package transform

import (
	"os"
	"reflect"
	"testing"

	"kasho/proto"
	"gopkg.in/yaml.v3"
)

func TestGetFakeValue(t *testing.T) {
	config := &Config{
		Version: ConfigVersionV1,
		Tables: map[string]TableConfig{
			"users": {
				"name":  {Type: Name},
				"age":   {Type: Year},
				"email": {Type: Email},
			},
		},
	}

	tests := []struct {
		name     string
		table    string
		column   string
		original *proto.ColumnValue
		want     *proto.ColumnValue
		wantErr  bool
	}{
		{
			name:   "transform name",
			table:  "users",
			column: "name",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "Lucy Welch"},
			},
			wantErr: false,
		},
		{
			name:   "transform age",
			table:  "users",
			column: "age",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 30},
			},
			want: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 1906},
			},
			wantErr: false,
		},
		{
			name:   "transform email",
			table:  "users",
			column: "email",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"},
			},
			want: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "zechariahkris@hackett.name"},
			},
			wantErr: false,
		},
		{
			name:   "no transform for unknown table",
			table:  "unknown",
			column: "name",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "no transform for unknown column",
			table:  "users",
			column: "unknown",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "type mismatch",
			table:  "users",
			column: "name",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 42},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFakeValue(config, tt.table, tt.column, tt.original)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFakeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFakeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTransformFunction(t *testing.T) {
	tests := []struct {
		name      string
		transform TransformType
		wantError bool
	}{
		{
			name:      "valid transform",
			transform: Name,
		},
		{
			name:      "valid transform",
			transform: Email,
		},
		{
			name:      "invalid transform",
			transform: "InvalidTransform",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.transform.GetTransformFunction()
			if (err != nil) != tt.wantError {
				t.Errorf("GetTransformFunction() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err != nil {
				return
			}

			if got == nil {
				t.Error("GetTransformFunction() returned nil function")
			}
		})
	}
}

func TestValidateAndMigrateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid v1 config",
			config: &Config{
				Version: ConfigVersionV1,
				Tables: map[string]TableConfig{
					"users": {"name": {Type: Name}},
				},
			},
			wantError: false,
		},
		{
			name: "config without version (legacy)",
			config: &Config{
				Tables: map[string]TableConfig{
					"users": {"name": {Type: Name}},
				},
			},
			wantError: false,
		},
		{
			name: "unsupported version",
			config: &Config{
				Version: "v2",
				Tables: map[string]TableConfig{
					"users": {"name": {Type: Name}},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndMigrateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAndMigrateConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// For legacy configs, ensure version was set to v1
			if tt.name == "config without version (legacy)" && tt.config.Version != ConfigVersionV1 {
				t.Errorf("Expected version to be set to %s for legacy config, got %s", ConfigVersionV1, tt.config.Version)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tmpDir := t.TempDir()
	
	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "valid v1 config file",
			content: `version: v1
tables:
  users:
    name: Name
    email: Email`,
			wantError: false,
		},
		{
			name: "legacy config file without version",
			content: `tables:
  users:
    name: Name`,
			wantError: false,
		},
		{
			name: "invalid yaml",
			content: `version: v1
tables:
  users
    name: Name`,
			wantError: true,
		},
		{
			name: "unsupported version",
			content: `version: v2
tables:
  users:
    name: Name`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test config to temporary file
			configPath := tmpDir + "/test_config.yaml"
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Test LoadConfig
			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if config == nil {
					t.Error("LoadConfig() returned nil config for valid input")
					return
				}
				if config.Version == "" {
					t.Error("LoadConfig() returned config with empty version")
				}
				if config.Tables == nil {
					t.Error("LoadConfig() returned config with nil tables")
				}
			}
		})
	}

	// Test file not found error
	t.Run("file not found", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/config.yaml")
		if err == nil {
			t.Error("LoadConfig() should return error for nonexistent file")
		}
	})
}

func TestGetFakeValueExtended(t *testing.T) {
	config := &Config{
		Version: ConfigVersionV1,
		Tables: map[string]TableConfig{
			"users": {
				"name":      {Type: Name},        // string->string
				"age":       {Type: Year},        // int->int (Year transform takes int, returns int)
				"balance":   {Type: Currency},    // string->string (Currency is string transform)
				"active":    {Type: Bool},        // bool->bool (Bool transform takes bool, returns bool)
				"latitude":  {Type: Latitude},    // float64->float64 (Latitude transform takes float64, returns float64)
				"timestamp": {Type: DateOfBirth}, // string->string (DateOfBirth is string transform)
			},
		},
	}

	tests := []struct {
		name        string
		table       string
		column      string
		original    *proto.ColumnValue
		expectError bool
	}{
		{
			name:   "int to int transformation (age)",
			table:  "users",
			column: "age",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 30},
			},
			expectError: false,
		},
		{
			name:   "bool to bool transformation (active)",
			table:  "users",
			column: "active",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_BoolValue{BoolValue: true},
			},
			expectError: false,
		},
		{
			name:   "float to float transformation (latitude)",
			table:  "users",
			column: "latitude",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_FloatValue{FloatValue: 40.7128},
			},
			expectError: false,
		},
		{
			name:   "string to string transformation (currency)",
			table:  "users",
			column: "balance",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "100.00"},
			},
			expectError: false,
		},
		{
			name:   "timestamp value as string (parsed to time.Time causes error)",
			table:  "users",
			column: "timestamp",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2023-01-01T00:00:00Z"},
			},
			expectError: true, // Valid timestamp gets parsed to time.Time, but DateOfBirth expects string
		},
		{
			name:   "invalid timestamp format",
			table:  "users",
			column: "timestamp",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_TimestampValue{TimestampValue: "invalid-timestamp"},
			},
			expectError: false, // Should still work, just uses string value
		},
		{
			name:   "type mismatch - int to string transform",
			table:  "users",
			column: "name",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 42},
			},
			expectError: true,
		},
		{
			name:   "unsupported value type",
			table:  "users",
			column: "name",
			original: &proto.ColumnValue{
				Value: nil, // This will cause an unsupported type error
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFakeValue(config, tt.table, tt.column, tt.original)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestTransformChange(t *testing.T) {
	config := &Config{
		Version: ConfigVersionV1,
		Tables: map[string]TableConfig{
			"public.users": {
				"name":  {Type: Name},
				"email": {Type: Email},
			},
		},
	}

	tests := []struct {
		name        string
		change      *proto.Change
		expectError bool
	}{
		{
			name: "DML INSERT change",
			change: &proto.Change{
				Lsn:  "123",
				Type: "DML",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "public.users",
						ColumnNames: []string{"id", "name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "INSERT",
					},
				},
			},
			expectError: false,
		},
		{
			name: "DML UPDATE change with old keys",
			change: &proto.Change{
				Lsn:  "124",
				Type: "DML",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "public.users",
						ColumnNames: []string{"id", "name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "Jane Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "jane@example.com"}},
						},
						Kind: "UPDATE",
						OldKeys: &proto.OldKeys{
							KeyNames:  []string{"id"},
							KeyValues: []*proto.ColumnValue{{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "DDL change",
			change: &proto.Change{
				Lsn:  "125",
				Type: "DDL",
				Data: &proto.Change_Ddl{
					Ddl: &proto.DDLData{
						Ddl: "CREATE TABLE test (id INT PRIMARY KEY)",
					},
				},
			},
			expectError: false,
		},
		{
			name: "unknown table (no transform)",
			change: &proto.Change{
				Lsn:  "126",
				Type: "DML",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "unknown.table",
						ColumnNames: []string{"id", "name"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "John"}},
						},
						Kind: "INSERT",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TransformChange(config, tt.change)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}

				// Verify the change structure is preserved
				if result.Lsn != tt.change.Lsn {
					t.Errorf("LSN mismatch: got %s, want %s", result.Lsn, tt.change.Lsn)
				}
				if result.Type != tt.change.Type {
					t.Errorf("Type mismatch: got %v, want %v", result.Type, tt.change.Type)
				}

				// For DML changes, verify data is properly copied
				if dmlData := result.GetDml(); dmlData != nil {
					originalDML := tt.change.GetDml()
					if dmlData.Table != originalDML.Table {
						t.Errorf("Table mismatch: got %s, want %s", dmlData.Table, originalDML.Table)
					}
					if len(dmlData.ColumnNames) != len(originalDML.ColumnNames) {
						t.Errorf("Column names length mismatch: got %d, want %d", 
							len(dmlData.ColumnNames), len(originalDML.ColumnNames))
					}
					if len(dmlData.ColumnValues) != len(originalDML.ColumnValues) {
						t.Errorf("Column values length mismatch: got %d, want %d", 
							len(dmlData.ColumnValues), len(originalDML.ColumnValues))
					}
				}
			}
		})
	}
}

func TestRegexTransform(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		replacement string
		input       string
		want        string
		wantErr     bool
	}{
		{
			name:        "phone number masking",
			pattern:     `\(\d{3}\) \d{3}-\d{4}`,
			replacement: "(XXX) XXX-XXXX",
			input:       "(123) 456-7890",
			want:        "(XXX) XXX-XXXX",
		},
		{
			name:        "credit card partial masking",
			pattern:     `(\d{4})-(\d{4})-(\d{4})-(\d{4})`,
			replacement: "XXXX-XXXX-XXXX-$4",
			input:       "1234-5678-9012-3456",
			want:        "XXXX-XXXX-XXXX-3456",
		},
		{
			name:        "email domain replacement",
			pattern:     `@[\w.-]+\.[\w.-]+`,
			replacement: "@example.com",
			input:       "john.doe@company.org",
			want:        "john.doe@example.com",
		},
		{
			name:        "IP address masking",
			pattern:     `\d+\.\d+\.\d+\.\d+`,
			replacement: "XXX.XXX.XXX.XXX",
			input:       "192.168.1.100",
			want:        "XXX.XXX.XXX.XXX",
		},
		{
			name:        "invalid regex pattern",
			pattern:     `[`,
			replacement: "replacement",
			input:       "test",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformFunc := TransformRegex(tt.pattern, tt.replacement)
			got, err := transformFunc(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("TransformRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFakeValueWithRegex(t *testing.T) {
	config := &Config{
		Version: ConfigVersionV1,
		Tables: map[string]TableConfig{
			"users": {
				"phone": {
					Type:        Regex,
					Pattern:     `\d{3}-\d{3}-\d{4}`,
					Replacement: "XXX-XXX-XXXX",
				},
				"ssn": {
					Type:        Regex,
					Pattern:     `(\d{3})-(\d{2})-(\d{4})`,
					Replacement: "XXX-XX-$3",
				},
			},
		},
	}

	tests := []struct {
		name     string
		table    string
		column   string
		original *proto.ColumnValue
		want     *proto.ColumnValue
		wantErr  bool
	}{
		{
			name:   "regex transform phone",
			table:  "users",
			column: "phone",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "123-456-7890"},
			},
			want: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "XXX-XXX-XXXX"},
			},
		},
		{
			name:   "regex transform ssn with capture group",
			table:  "users",
			column: "ssn",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "123-45-6789"},
			},
			want: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "XXX-XX-6789"},
			},
		},
		{
			name:   "regex transform on non-string value",
			table:  "users",
			column: "phone",
			original: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 1234567890},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFakeValue(config, tt.table, tt.column, tt.original)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFakeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFakeValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnTransformUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    TableConfig
		wantErr bool
	}{
		{
			name: "simple string format",
			yaml: `
name: Name
email: Email
phone: Phone`,
			want: TableConfig{
				"name":  {Type: Name},
				"email": {Type: Email},
				"phone": {Type: Phone},
			},
		},
		{
			name: "object format",
			yaml: `
name:
  type: Name
email:
  type: Email
phone:
  type: Regex
  pattern: '\d{3}-\d{3}-\d{4}'
  replacement: 'XXX-XXX-XXXX'`,
			want: TableConfig{
				"name":  {Type: Name},
				"email": {Type: Email},
				"phone": {Type: Regex, Pattern: `\d{3}-\d{3}-\d{4}`, Replacement: "XXX-XXX-XXXX"},
			},
		},
		{
			name: "mixed format",
			yaml: `
name: Name
phone:
  type: Regex
  pattern: '\d+'
  replacement: 'XXX'
email: Email`,
			want: TableConfig{
				"name":  {Type: Name},
				"phone": {Type: Regex, Pattern: `\d+`, Replacement: "XXX"},
				"email": {Type: Email},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TableConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}
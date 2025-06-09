package transform

import (
	"reflect"
	"testing"

	"kasho/proto"
)

func TestGetFakeValue(t *testing.T) {
	config := &Config{
		Version: ConfigVersionV1,
		Tables: map[string]TableConfig{
			"users": {
				"name":  Name,
				"age":   Year,
				"email": Email,
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
					"users": {"name": Name},
				},
			},
			wantError: false,
		},
		{
			name: "config without version (legacy)",
			config: &Config{
				Tables: map[string]TableConfig{
					"users": {"name": Name},
				},
			},
			wantError: false,
		},
		{
			name: "unsupported version",
			config: &Config{
				Version: "v2",
				Tables: map[string]TableConfig{
					"users": {"name": Name},
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

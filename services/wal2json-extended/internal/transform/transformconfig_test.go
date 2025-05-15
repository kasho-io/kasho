package main

import (
	"testing"
	"time"
)

func TestGetFakeValue(t *testing.T) {
	config := &Config{
		Tables: map[string]TableConfig{
			"users": {
				"name":     FullName,
				"age":      Year,
				"email":    Email,
				"is_admin": Bool,
			},
		},
	}

	tests := []struct {
		name      string
		table     string
		column    string
		wantType  string
		wantError bool
	}{
		{
			name:     "valid string column",
			table:    "users",
			column:   "name",
			wantType: "string",
		},
		{
			name:     "valid int column",
			table:    "users",
			column:   "age",
			wantType: "int",
		},
		{
			name:     "valid email column",
			table:    "users",
			column:   "email",
			wantType: "string",
		},
		{
			name:      "invalid table",
			table:     "nonexistent",
			column:    "name",
			wantError: true,
		},
		{
			name:      "invalid column",
			table:     "users",
			column:    "nonexistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.GetFakeValue(tt.table, tt.column)
			if (err != nil) != tt.wantError {
				t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err != nil {
				return
			}

			switch tt.wantType {
			case "string":
				if _, ok := got.(string); !ok {
					t.Errorf("GetFakeValue() = %T, want string", got)
				}
			case "int":
				if _, ok := got.(int); !ok {
					t.Errorf("GetFakeValue() = %T, want int", got)
				}
			case "float64":
				if _, ok := got.(float64); !ok {
					t.Errorf("GetFakeValue() = %T, want float64", got)
				}
			case "bool":
				if _, ok := got.(bool); !ok {
					t.Errorf("GetFakeValue() = %T, want bool", got)
				}
			case "time.Time":
				if _, ok := got.(time.Time); !ok {
					t.Errorf("GetFakeValue() = %T, want time.Time", got)
				}
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
			transform: FullName,
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

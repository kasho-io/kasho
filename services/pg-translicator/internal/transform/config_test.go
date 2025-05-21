package transform

import (
	"testing"
	"time"
)

func TestGetFakeValue(t *testing.T) {
	config := &Config{
		Tables: map[string]TableConfig{
			"users": {
				"name":     Name,
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
		original  any
		wantType  string
		wantError bool
	}{
		{
			name:     "valid string column",
			table:    "users",
			column:   "name",
			original: "test123",
			wantType: "string",
		},
		{
			name:     "valid int column",
			table:    "users",
			column:   "age",
			original: 25,
			wantType: "int",
		},
		{
			name:     "valid email column",
			table:    "users",
			column:   "email",
			original: "test123",
			wantType: "string",
		},
		{
			name:      "unmapped table",
			table:     "nonexistent",
			column:    "name",
			original:  "test123",
			wantError: false,
		},
		{
			name:      "unmapped column",
			table:     "users",
			column:    "nonexistent",
			original:  "test123",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.original.(type) {
			case string:
				got, err := GetFakeValue(config, tt.table, tt.column, v)
				if (err != nil) != tt.wantError {
					t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
					return
				}
				if err != nil {
					return
				}
				checkResult(t, got, tt.wantType)
			case int:
				got, err := GetFakeValue(config, tt.table, tt.column, v)
				if (err != nil) != tt.wantError {
					t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
					return
				}
				if err != nil {
					return
				}
				checkResult(t, got, tt.wantType)
			case float64:
				got, err := GetFakeValue(config, tt.table, tt.column, v)
				if (err != nil) != tt.wantError {
					t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
					return
				}
				if err != nil {
					return
				}
				checkResult(t, got, tt.wantType)
			case bool:
				got, err := GetFakeValue(config, tt.table, tt.column, v)
				if (err != nil) != tt.wantError {
					t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
					return
				}
				if err != nil {
					return
				}
				checkResult(t, got, tt.wantType)
			case time.Time:
				got, err := GetFakeValue(config, tt.table, tt.column, v)
				if (err != nil) != tt.wantError {
					t.Errorf("GetFakeValue() error = %v, wantError %v", err, tt.wantError)
					return
				}
				if err != nil {
					return
				}
				checkResult(t, got, tt.wantType)
			default:
				t.Errorf("unsupported type: %T", tt.original)
			}
		})
	}
}

func checkResult(t *testing.T, got any, wantType string) {
	switch wantType {
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

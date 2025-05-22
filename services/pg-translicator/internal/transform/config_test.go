package transform

import (
	"reflect"
	"testing"

	"pg-change-stream/api"
)

func TestGetFakeValue(t *testing.T) {
	config := &Config{
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
		original *api.ColumnValue
		want     *api.ColumnValue
		wantErr  bool
	}{
		{
			name:   "transform name",
			table:  "users",
			column: "name",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "Lucy Welch"},
			},
			wantErr: false,
		},
		{
			name:   "transform age",
			table:  "users",
			column: "age",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_IntValue{IntValue: 30},
			},
			want: &api.ColumnValue{
				Value: &api.ColumnValue_IntValue{IntValue: 1906},
			},
			wantErr: false,
		},
		{
			name:   "transform email",
			table:  "users",
			column: "email",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "john@example.com"},
			},
			want: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "zechariahkris@hackett.name"},
			},
			wantErr: false,
		},
		{
			name:   "no transform for unknown table",
			table:  "unknown",
			column: "name",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "no transform for unknown column",
			table:  "users",
			column: "unknown",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_StringValue{StringValue: "John Doe"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "type mismatch",
			table:  "users",
			column: "name",
			original: &api.ColumnValue{
				Value: &api.ColumnValue_IntValue{IntValue: 42},
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

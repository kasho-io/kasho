package types

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"kasho/proto"
)

func TestColumnValueWrapper_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		cv       ColumnValueWrapper
		wantJSON string
		wantErr  bool
	}{
		{
			name: "string value",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_StringValue{StringValue: "hello"},
				},
			},
			wantJSON: `"hello"`,
		},
		{
			name: "int value",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_IntValue{IntValue: 42},
				},
			},
			wantJSON: "42",
		},
		{
			name: "float value",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_FloatValue{FloatValue: 3.14},
				},
			},
			wantJSON: "3.14",
		},
		{
			name: "bool value true",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_BoolValue{BoolValue: true},
				},
			},
			wantJSON: "true",
		},
		{
			name: "bool value false",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_BoolValue{BoolValue: false},
				},
			},
			wantJSON: "false",
		},
		{
			name: "timestamp value",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"},
				},
			},
			wantJSON: `"2024-03-20T15:04:05Z"`,
		},
		{
			name: "nil column value",
			cv: ColumnValueWrapper{
				ColumnValue: nil,
			},
			wantJSON: "null",
		},
		{
			name: "unset oneof value (SQL NULL)",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{},
			},
			wantJSON: "null",
		},
		{
			name: "string with special characters",
			cv: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_StringValue{StringValue: "hello\nworld\t\"quoted\""},
				},
			},
			wantJSON: `"hello\nworld\t\"quoted\""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cv.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ColumnValueWrapper.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.wantJSON {
				t.Errorf("ColumnValueWrapper.MarshalJSON() = %v, want %v", string(got), tt.wantJSON)
			}
		})
	}
}

func TestColumnValueWrapper_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     ColumnValueWrapper
		wantErr  bool
	}{
		{
			name:     "string value",
			jsonData: `"hello"`,
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_StringValue{StringValue: "hello"},
				},
			},
		},
		{
			name:     "int value",
			jsonData: "42",
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_IntValue{IntValue: 42},
				},
			},
		},
		{
			name:     "float value",
			jsonData: "3.14",
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_FloatValue{FloatValue: 3.14},
				},
			},
		},
		{
			name:     "bool value true",
			jsonData: "true",
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_BoolValue{BoolValue: true},
				},
			},
		},
		{
			name:     "bool value false",
			jsonData: "false",
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_BoolValue{BoolValue: false},
				},
			},
		},
		{
			name:     "null value (SQL NULL)",
			jsonData: "null",
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: nil, // Unset oneof represents NULL
				},
			},
		},
		{
			name:     "timestamp value",
			jsonData: `"2024-03-20T15:04:05Z"`,
			want: ColumnValueWrapper{
				ColumnValue: &proto.ColumnValue{
					Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"},
				},
			},
		},
		{
			name:     "invalid json",
			jsonData: `{"invalid": json}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cv ColumnValueWrapper
			err := cv.UnmarshalJSON([]byte(tt.jsonData))
			if (err != nil) != tt.wantErr {
				t.Errorf("ColumnValueWrapper.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(cv, tt.want) {
				t.Errorf("ColumnValueWrapper.UnmarshalJSON() = %v, want %v", cv, tt.want)
			}
		})
	}
}

func TestChange_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		change   Change
		wantJSON string
	}{
		{
			name: "DML change",
			change: Change{
				Position: "0/100",
				Data: &DMLData{
					Table:       "users",
					Kind:        "insert",
					ColumnNames: []string{"id", "name"},
					ColumnValues: []ColumnValueWrapper{
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "test"}}},
					},
				},
			},
			wantJSON: `{"type":"dml","position":"0/100","data":{"table":"users","columnnames":["id","name"],"columnvalues":[1,"test"],"kind":"insert"}}`,
		},
		{
			name: "DDL change",
			change: Change{
				Position: "0/200",
				Data: &DDLData{
					ID:       1,
					Time:     time.Date(2024, 3, 20, 15, 0, 0, 0, time.UTC),
					Username: "postgres",
					Database: "testdb",
					DDL:      "CREATE TABLE test (id SERIAL PRIMARY KEY)",
				},
			},
			wantJSON: `{"type":"ddl","position":"0/200","data":{"id":1,"time":"2024-03-20T15:00:00Z","username":"postgres","database":"testdb","ddl":"CREATE TABLE test (id SERIAL PRIMARY KEY)"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.change.MarshalJSON()
			if err != nil {
				t.Errorf("Change.MarshalJSON() error = %v", err)
				return
			}
			if string(got) != tt.wantJSON {
				t.Errorf("Change.MarshalJSON() = %v, want %v", string(got), tt.wantJSON)
			}
		})
	}
}

func TestChange_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		want     Change
		wantErr  bool
	}{
		{
			name:     "DML change",
			jsonData: `{"type":"dml","position":"0/100","data":{"table":"users","columnnames":["id","name"],"columnvalues":[1,"test"],"kind":"insert"}}`,
			want: Change{
				Position: "0/100",
				Data: &DMLData{
					Table:       "users",
					Kind:        "insert",
					ColumnNames: []string{"id", "name"},
					ColumnValues: []ColumnValueWrapper{
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "test"}}},
					},
				},
			},
		},
		{
			name:     "DDL change",
			jsonData: `{"type":"ddl","position":"0/200","data":{"id":1,"time":"2024-03-20T15:00:00Z","username":"postgres","database":"testdb","ddl":"CREATE TABLE test (id SERIAL PRIMARY KEY)"}}`,
			want: Change{
				Position: "0/200",
				Data: &DDLData{
					ID:       1,
					Time:     time.Date(2024, 3, 20, 15, 0, 0, 0, time.UTC),
					Username: "postgres",
					Database: "testdb",
					DDL:      "CREATE TABLE test (id SERIAL PRIMARY KEY)",
				},
			},
		},
		{
			name:     "unknown change type",
			jsonData: `{"type":"unknown","position":"0/100","data":{}}`,
			wantErr:  true,
		},
		{
			name:     "invalid json",
			jsonData: `{"type":"dml"`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var change Change
			err := change.UnmarshalJSON([]byte(tt.jsonData))
			if (err != nil) != tt.wantErr {
				t.Errorf("Change.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(change, tt.want) {
				t.Errorf("Change.UnmarshalJSON() = %v, want %v", change, tt.want)
			}
		})
	}
}

func TestRoundTripSerialization(t *testing.T) {
	// Test that marshaling and unmarshaling produces the same result
	original := Change{
		Position: "0/12345",
		Data: &DMLData{
			Table:       "products",
			Kind:        "update",
			ColumnNames: []string{"name", "price", "active"},
			ColumnValues: []ColumnValueWrapper{
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "Widget"}}},
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: 19.99}}},
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: true}}},
			},
			OldKeys: &struct {
				KeyNames  []string             `json:"keynames"`
				KeyValues []ColumnValueWrapper `json:"keyvalues"`
			}{
				KeyNames: []string{"id"},
				KeyValues: []ColumnValueWrapper{
					{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 123}}},
				},
			},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var decoded Change
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare
	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Round trip failed: original = %+v, decoded = %+v", original, decoded)
	}
}

func TestDMLData_Type(t *testing.T) {
	dml := &DMLData{}
	if dml.Type() != "dml" {
		t.Errorf("DMLData.Type() = %v, want dml", dml.Type())
	}
}

func TestDDLData_Type(t *testing.T) {
	ddl := &DDLData{}
	if ddl.Type() != "ddl" {
		t.Errorf("DDLData.Type() = %v, want ddl", ddl.Type())
	}
}

func TestChange_Type(t *testing.T) {
	tests := []struct {
		name   string
		change Change
		want   string
	}{
		{
			name: "DML change",
			change: Change{
				Data: &DMLData{},
			},
			want: "dml",
		},
		{
			name: "DDL change",
			change: Change{
				Data: &DDLData{},
			},
			want: "ddl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.change.Type(); got != tt.want {
				t.Errorf("Change.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
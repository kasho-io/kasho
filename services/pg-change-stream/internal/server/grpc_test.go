package server

import (
	"reflect"
	"testing"
	"time"

	"kasho/proto"
	"pg-change-stream/internal/types"
)

func TestConvertToProtoChange_DMLData(t *testing.T) {
	tests := []struct {
		name   string
		change types.Change
		want   *proto.Change
	}{
		{
			name: "DML insert without old keys",
			change: types.Change{
				LSN: "0/100",
				Data: &types.DMLData{
					Table:       "public.users",
					Kind:        "insert",
					ColumnNames: []string{"id", "name", "email"},
					ColumnValues: []types.ColumnValueWrapper{
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}}},
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}}},
					},
				},
			},
			want: &proto.Change{
				Lsn:  "0/100",
				Type: "dml",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "public.users",
						Kind:        "insert",
						ColumnNames: []string{"id", "name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						OldKeys: nil,
					},
				},
			},
		},
		{
			name: "DML update with old keys",
			change: types.Change{
				LSN: "0/200",
				Data: &types.DMLData{
					Table:       "public.users",
					Kind:        "update",
					ColumnNames: []string{"name"},
					ColumnValues: []types.ColumnValueWrapper{
						{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "Jane Doe"}}},
					},
					OldKeys: &struct {
						KeyNames  []string                   `json:"keynames"`
						KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
					}{
						KeyNames: []string{"id"},
						KeyValues: []types.ColumnValueWrapper{
							{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
						},
					},
				},
			},
			want: &proto.Change{
				Lsn:  "0/200",
				Type: "dml",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "public.users",
						Kind:        "update",
						ColumnNames: []string{"name"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "Jane Doe"}},
						},
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
							},
						},
					},
				},
			},
		},
		{
			name: "DML delete with old keys",
			change: types.Change{
				LSN: "0/300",
				Data: &types.DMLData{
					Table:        "public.users",
					Kind:         "delete",
					ColumnNames:  []string{},
					ColumnValues: []types.ColumnValueWrapper{},
					OldKeys: &struct {
						KeyNames  []string                   `json:"keynames"`
						KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
					}{
						KeyNames: []string{"id", "email"},
						KeyValues: []types.ColumnValueWrapper{
							{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
							{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}}},
						},
					},
				},
			},
			want: &proto.Change{
				Lsn:  "0/300",
				Type: "dml",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:        "public.users",
						Kind:         "delete",
						ColumnNames:  []string{},
						ColumnValues: []*proto.ColumnValue{},
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"id", "email"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
								{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToProtoChange(tt.change)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToProtoChange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToProtoChange_DDLData(t *testing.T) {
	testTime := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)
	
	change := types.Change{
		LSN: "0/400",
		Data: &types.DDLData{
			ID:       123,
			Time:     testTime,
			Username: "postgres",
			Database: "testdb",
			DDL:      "CREATE TABLE test (id SERIAL PRIMARY KEY, name VARCHAR(100))",
		},
	}

	want := &proto.Change{
		Lsn:  "0/400",
		Type: "ddl",
		Data: &proto.Change_Ddl{
			Ddl: &proto.DDLData{
				Id:       123,
				Time:     testTime.Format(time.RFC3339),
				Username: "postgres",
				Database: "testdb",
				Ddl:      "CREATE TABLE test (id SERIAL PRIMARY KEY, name VARCHAR(100))",
			},
		},
	}

	got := convertToProtoChange(change)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("convertToProtoChange() = %v, want %v", got, want)
	}
}

func TestConvertToProtoChange_DifferentColumnTypes(t *testing.T) {
	change := types.Change{
		LSN: "0/500",
		Data: &types.DMLData{
			Table:       "public.test_table",
			Kind:        "insert",
			ColumnNames: []string{"bool_col", "float_col", "timestamp_col"},
			ColumnValues: []types.ColumnValueWrapper{
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: true}}},
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: 3.14}}},
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"}}},
			},
		},
	}

	want := &proto.Change{
		Lsn:  "0/500",
		Type: "dml",
		Data: &proto.Change_Dml{
			Dml: &proto.DMLData{
				Table:       "public.test_table",
				Kind:        "insert",
				ColumnNames: []string{"bool_col", "float_col", "timestamp_col"},
				ColumnValues: []*proto.ColumnValue{
					{Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
					{Value: &proto.ColumnValue_FloatValue{FloatValue: 3.14}},
					{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"}},
				},
				OldKeys: nil,
			},
		},
	}

	got := convertToProtoChange(change)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("convertToProtoChange() = %v, want %v", got, want)
	}
}

func TestConvertToProtoChange_EmptyData(t *testing.T) {
	change := types.Change{
		LSN: "0/600",
		Data: &types.DMLData{
			Table:        "public.empty_table",
			Kind:         "insert",
			ColumnNames:  []string{},
			ColumnValues: []types.ColumnValueWrapper{},
		},
	}

	want := &proto.Change{
		Lsn:  "0/600",
		Type: "dml",
		Data: &proto.Change_Dml{
			Dml: &proto.DMLData{
				Table:        "public.empty_table",
				Kind:         "insert",
				ColumnNames:  []string{},
				ColumnValues: []*proto.ColumnValue{},
				OldKeys:      nil,
			},
		},
	}

	got := convertToProtoChange(change)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("convertToProtoChange() = %v, want %v", got, want)
	}
}
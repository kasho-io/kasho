package sql

import (
	"testing"

	"pg-change-stream/api"
)

func TestToSQL(t *testing.T) {
	tests := []struct {
		name    string
		change  *api.Change
		wantSQL string
		wantErr bool
	}{
		{
			name: "valid insert",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com');",
			wantErr: false,
		},
		{
			name: "valid update",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "update",
						OldKeys: &api.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*api.ColumnValue{
								{Value: &api.ColumnValue_IntValue{IntValue: 1}},
							},
						},
					},
				},
			},
			wantSQL: "UPDATE users SET name = 'John Doe', email = 'john@example.com' WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "valid delete",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table: "users",
						Kind:  "delete",
						OldKeys: &api.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*api.ColumnValue{
								{Value: &api.ColumnValue_IntValue{IntValue: 1}},
							},
						},
					},
				},
			},
			wantSQL: "DELETE FROM users WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "valid DDL",
			change: &api.Change{
				Data: &api.Change_Ddl{
					Ddl: &api.DDLData{
						Ddl: "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT);",
					},
				},
			},
			wantSQL: "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT);",
			wantErr: false,
		},
		{
			name: "invalid DML kind",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table: "users",
						Kind:  "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mismatched columns and values",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "insert",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update without old keys",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
						},
						Kind: "update",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "delete without old keys",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table: "users",
						Kind:  "delete",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "insert with null value",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: nil},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, email) VALUES ('John Doe', NULL);",
			wantErr: false,
		},
		{
			name: "insert with number",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "age"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_IntValue{IntValue: 42}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, age) VALUES ('John Doe', 42);",
			wantErr: false,
		},
		{
			name: "insert with boolean",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "is_active"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_BoolValue{BoolValue: true}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, is_active) VALUES ('John Doe', true);",
			wantErr: false,
		},
		{
			name: "insert with timestamp",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "created_at"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, created_at) VALUES ('John Doe', '2024-03-20 15:04:05');",
			wantErr: false,
		},
		{
			name: "insert with date",
			change: &api.Change{
				Data: &api.Change_Dml{
					Dml: &api.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "birth_date"},
						ColumnValues: []*api.ColumnValue{
							{Value: &api.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &api.ColumnValue_TimestampValue{TimestampValue: "2024-03-20"}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, birth_date) VALUES ('John Doe', '2024-03-20');",
			wantErr: false,
		},
		{
			name: "unsupported change type",
			change: &api.Change{
				Data: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToSQL(tt.change)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.wantSQL {
				t.Errorf("ToSQL() = %v, want %v", got, tt.wantSQL)
			}
		})
	}
}

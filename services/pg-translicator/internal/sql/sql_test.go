package sql

import (
	"testing"

	"kasho/proto"
)

func TestToSQL(t *testing.T) {
	tests := []struct {
		name    string
		change  *proto.Change
		wantSQL string
		wantErr bool
	}{
		{
			name: "valid insert",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "update",
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table: "users",
						Kind:  "delete",
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 1}},
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
			change: &proto.Change{
				Data: &proto.Change_Ddl{
					Ddl: &proto.DDLData{
						Ddl: "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT);",
					},
				},
			},
			wantSQL: "CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT);",
			wantErr: false,
		},
		{
			name: "invalid DML kind",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table: "users",
						Kind:  "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "mismatched columns and values",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "john@example.com"}},
						},
						Kind: "insert",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update without old keys",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
						},
						Kind: "update",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "delete without old keys",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table: "users",
						Kind:  "delete",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "insert with null value",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "email"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "age"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_IntValue{IntValue: 42}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "is_active"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "created_at"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20T15:04:05Z"}},
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
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "birth_date"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
							{Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-03-20"}},
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
			change: &proto.Change{
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

// TestToSQL_ComplexScenarios tests more complex SQL generation scenarios
func TestToSQL_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name    string
		change  *proto.Change
		wantSQL string
		wantErr bool
	}{
		{
			name: "update with composite primary key",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "user_roles",
						ColumnNames: []string{"role_name", "permissions"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "admin"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "read,write,delete"}},
						},
						Kind: "update",
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"user_id", "org_id"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 123}},
								{Value: &proto.ColumnValue_IntValue{IntValue: 456}},
							},
						},
					},
				},
			},
			wantSQL: "UPDATE user_roles SET role_name = 'admin', permissions = 'read,write,delete' WHERE user_id = 123 AND org_id = 456;",
			wantErr: false,
		},
		{
			name: "string with single quotes",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "users",
						ColumnNames: []string{"name", "bio"},
						ColumnValues: []*proto.ColumnValue{
							{Value: &proto.ColumnValue_StringValue{StringValue: "O'Connor"}},
							{Value: &proto.ColumnValue_StringValue{StringValue: "He said 'Hello World'"}},
						},
						Kind: "insert",
					},
				},
			},
			wantSQL: "INSERT INTO users (name, bio) VALUES ('O''Connor', 'He said ''Hello World''');",
			wantErr: false,
		},
		{
			name: "unknown DML kind error",
			change: &proto.Change{
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table: "users",
						Kind:  "merge", // unsupported operation
					},
				},
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

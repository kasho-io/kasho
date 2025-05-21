package sql

import (
	"testing"

	"pg-change-stream/api"
)

func TestToSQL(t *testing.T) {
	tests := []struct {
		name    string
		dml     *api.DMLData
		wantSQL string
		wantErr bool
	}{
		{
			name: "valid insert",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "email"},
				ColumnValues: []string{"John Doe", "john@example.com"},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com');",
			wantErr: false,
		},
		{
			name: "valid update",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "email"},
				ColumnValues: []string{"John Doe", "john@example.com"},
				Kind:         "update",
				OldKeys: &api.OldKeys{
					KeyNames:  []string{"id"},
					KeyValues: []string{"1"},
				},
			},
			wantSQL: "UPDATE users SET name = 'John Doe', email = 'john@example.com' WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "valid delete",
			dml: &api.DMLData{
				Table: "users",
				Kind:  "delete",
				OldKeys: &api.OldKeys{
					KeyNames:  []string{"id"},
					KeyValues: []string{"1"},
				},
			},
			wantSQL: "DELETE FROM users WHERE id = 1;",
			wantErr: false,
		},
		{
			name: "invalid kind",
			dml: &api.DMLData{
				Table: "users",
				Kind:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "mismatched columns and values",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name"},
				ColumnValues: []string{"John Doe", "john@example.com"},
				Kind:         "insert",
			},
			wantErr: true,
		},
		{
			name: "update without old keys",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name"},
				ColumnValues: []string{"John Doe"},
				Kind:         "update",
			},
			wantErr: true,
		},
		{
			name: "delete without old keys",
			dml: &api.DMLData{
				Table: "users",
				Kind:  "delete",
			},
			wantErr: true,
		},
		{
			name: "insert with null value",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "email"},
				ColumnValues: []string{"John Doe", ""},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, email) VALUES ('John Doe', NULL);",
			wantErr: false,
		},
		{
			name: "insert with number",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "age"},
				ColumnValues: []string{"John Doe", "42"},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, age) VALUES ('John Doe', 42);",
			wantErr: false,
		},
		{
			name: "insert with boolean",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "is_active"},
				ColumnValues: []string{"John Doe", "true"},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, is_active) VALUES ('John Doe', true);",
			wantErr: false,
		},
		{
			name: "insert with timestamp",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "created_at"},
				ColumnValues: []string{"John Doe", "2024-03-20T15:04:05Z"},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, created_at) VALUES ('John Doe', '2024-03-20 15:04:05');",
			wantErr: false,
		},
		{
			name: "insert with date",
			dml: &api.DMLData{
				Table:        "users",
				ColumnNames:  []string{"name", "birth_date"},
				ColumnValues: []string{"John Doe", "2024-03-20"},
				Kind:         "insert",
			},
			wantSQL: "INSERT INTO users (name, birth_date) VALUES ('John Doe', '2024-03-20');",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToSQL(tt.dml)
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

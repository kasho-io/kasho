package server

import (
	"testing"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-mysql-org/go-mysql/schema"

	"kasho/pkg/types"
	"kasho/proto"
)

func TestFormatBinlogPosition(t *testing.T) {
	tests := []struct {
		name string
		pos  mysql.Position
		want string
	}{
		{
			name: "standard position",
			pos:  mysql.Position{Name: "mysql-bin.000001", Pos: 4},
			want: "mysql-bin.000001:4",
		},
		{
			name: "larger position",
			pos:  mysql.Position{Name: "binlog.000123", Pos: 123456},
			want: "binlog.000123:123456",
		},
		{
			name: "zero position",
			pos:  mysql.Position{Name: "mysql-bin.000001", Pos: 0},
			want: "mysql-bin.000001:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBinlogPosition(tt.pos)
			if got != tt.want {
				t.Errorf("FormatBinlogPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBinlogPosition(t *testing.T) {
	tests := []struct {
		name     string
		position string
		want     mysql.Position
		wantErr  bool
	}{
		{
			name:     "standard position",
			position: "mysql-bin.000001:4",
			want:     mysql.Position{Name: "mysql-bin.000001", Pos: 4},
			wantErr:  false,
		},
		{
			name:     "larger position",
			position: "binlog.000123:123456",
			want:     mysql.Position{Name: "binlog.000123", Pos: 123456},
			wantErr:  false,
		},
		{
			name:     "empty position returns zero",
			position: "",
			want:     mysql.Position{},
			wantErr:  false,
		},
		{
			name:     "bootstrap position returns zero",
			position: "bootstrap",
			want:     mysql.Position{},
			wantErr:  false,
		},
		{
			name:     "invalid format - no colon",
			position: "mysql-bin.000001",
			want:     mysql.Position{},
			wantErr:  true,
		},
		{
			name:     "invalid format - non-numeric offset",
			position: "mysql-bin.000001:abc",
			want:     mysql.Position{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBinlogPosition(tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBinlogPosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseBinlogPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryEventToChange_DDL(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		schema  string
		wantNil bool
	}{
		{
			name:    "CREATE TABLE",
			query:   "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100))",
			schema:  "testdb",
			wantNil: false,
		},
		{
			name:    "ALTER TABLE",
			query:   "ALTER TABLE users ADD COLUMN email VARCHAR(255)",
			schema:  "testdb",
			wantNil: false,
		},
		{
			name:    "DROP TABLE",
			query:   "DROP TABLE IF EXISTS old_table",
			schema:  "testdb",
			wantNil: false,
		},
		{
			name:    "TRUNCATE TABLE",
			query:   "TRUNCATE TABLE users",
			schema:  "testdb",
			wantNil: false,
		},
		{
			name:    "RENAME TABLE",
			query:   "RENAME TABLE old_name TO new_name",
			schema:  "testdb",
			wantNil: false,
		},
		{
			name:    "SELECT should be ignored",
			query:   "SELECT * FROM users",
			schema:  "testdb",
			wantNil: true,
		},
		{
			name:    "INSERT should be ignored",
			query:   "INSERT INTO users VALUES (1, 'test')",
			schema:  "testdb",
			wantNil: true,
		},
		{
			name:    "UPDATE should be ignored",
			query:   "UPDATE users SET name = 'test' WHERE id = 1",
			schema:  "testdb",
			wantNil: true,
		},
		{
			name:    "DELETE should be ignored",
			query:   "DELETE FROM users WHERE id = 1",
			schema:  "testdb",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &replication.EventHeader{
				Timestamp: 1710950645, // 2024-03-20 15:04:05 UTC
			}
			event := &replication.QueryEvent{
				Query:  []byte(tt.query),
				Schema: []byte(tt.schema),
			}
			pos := mysql.Position{Name: "mysql-bin.000001", Pos: 1234}

			got := QueryEventToChange(header, event, pos)

			if tt.wantNil {
				if got != nil {
					t.Errorf("QueryEventToChange() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("QueryEventToChange() = nil, want non-nil")
					return
				}
				if got.Type() != "ddl" {
					t.Errorf("QueryEventToChange().Type() = %v, want ddl", got.Type())
				}
				if got.GetPosition() != "mysql-bin.000001:1234" {
					t.Errorf("QueryEventToChange().GetPosition() = %v, want mysql-bin.000001:1234", got.GetPosition())
				}
			}
		})
	}
}

func TestQueryEventToChange_UsesHeaderTimestamp(t *testing.T) {
	// Unix timestamp for 2024-03-20 15:04:05 UTC
	expectedTimestamp := uint32(1710950645)
	header := &replication.EventHeader{
		Timestamp: expectedTimestamp,
	}
	event := &replication.QueryEvent{
		Query:  []byte("CREATE TABLE test (id INT)"),
		Schema: []byte("testdb"),
	}
	pos := mysql.Position{Name: "mysql-bin.000001", Pos: 1234}

	got := QueryEventToChange(header, event, pos)

	if got == nil {
		t.Fatal("QueryEventToChange() returned nil")
	}

	ddl, ok := got.Data.(types.DDLData)
	if !ok {
		t.Fatal("Expected DDLData type")
	}

	expectedTime := time.Unix(int64(expectedTimestamp), 0)
	if !ddl.Time.Equal(expectedTime) {
		t.Errorf("DDL time = %v, want %v", ddl.Time, expectedTime)
	}
}

func TestToColumnValue(t *testing.T) {
	col := &schema.TableColumn{Name: "test", Type: schema.TYPE_STRING}

	tests := []struct {
		name  string
		value any
		check func(t *testing.T, cv types.ColumnValueWrapper)
	}{
		{
			name:  "string value",
			value: "hello",
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetStringValue() != "hello" {
					t.Errorf("expected string 'hello', got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "int64 value",
			value: int64(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "int32 value",
			value: int32(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "int16 value",
			value: int16(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "int8 value",
			value: int8(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "int value",
			value: int(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "uint64 value",
			value: uint64(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "uint32 value",
			value: uint32(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "uint16 value",
			value: uint16(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "uint8 value",
			value: uint8(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "uint value",
			value: uint(42),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetIntValue() != 42 {
					t.Errorf("expected int 42, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "float64 value",
			value: float64(3.14),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetFloatValue() != 3.14 {
					t.Errorf("expected float 3.14, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "float32 value",
			value: float32(3.14),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				// float32 to float64 conversion may have slight precision differences
				if cv.ColumnValue.GetFloatValue() < 3.13 || cv.ColumnValue.GetFloatValue() > 3.15 {
					t.Errorf("expected float ~3.14, got %v", cv.ColumnValue.GetFloatValue())
				}
			},
		},
		{
			name:  "bool true",
			value: true,
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if !cv.ColumnValue.GetBoolValue() {
					t.Errorf("expected bool true, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "bool false",
			value: false,
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetBoolValue() {
					t.Errorf("expected bool false, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "time.Time value",
			value: time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				expected := "2024-03-20T15:04:05Z"
				if cv.ColumnValue.GetTimestampValue() != expected {
					t.Errorf("expected timestamp %s, got %v", expected, cv.ColumnValue.GetTimestampValue())
				}
			},
		},
		{
			name:  "byte slice",
			value: []byte("binary data"),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetStringValue() != "binary data" {
					t.Errorf("expected string 'binary data', got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "nil value",
			value: nil,
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetStringValue() != "" {
					t.Errorf("expected empty string for nil, got %v", cv.ColumnValue.GetValue())
				}
			},
		},
		{
			name:  "unknown type (struct) converts to string",
			value: struct{ Name string }{Name: "test"},
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetStringValue() != "{test}" {
					t.Errorf("expected string '{test}', got %v", cv.ColumnValue.GetStringValue())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toColumnValue(tt.value, col)
			tt.check(t, got)
		})
	}
}

func TestIsPrimaryKey(t *testing.T) {
	table := &schema.Table{
		Name:   "users",
		Schema: "testdb",
		Columns: []schema.TableColumn{
			{Name: "id", Type: schema.TYPE_NUMBER},
			{Name: "name", Type: schema.TYPE_STRING},
			{Name: "email", Type: schema.TYPE_STRING},
		},
		PKColumns: []int{0}, // id is the primary key
	}

	tests := []struct {
		name     string
		colIndex int
		want     bool
	}{
		{name: "id is primary key", colIndex: 0, want: true},
		{name: "name is not primary key", colIndex: 1, want: false},
		{name: "email is not primary key", colIndex: 2, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &table.Columns[tt.colIndex]
			got := isPrimaryKey(col, table)
			if got != tt.want {
				t.Errorf("isPrimaryKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPrimaryKey_CompositePK(t *testing.T) {
	table := &schema.Table{
		Name:   "order_items",
		Schema: "testdb",
		Columns: []schema.TableColumn{
			{Name: "order_id", Type: schema.TYPE_NUMBER},
			{Name: "item_id", Type: schema.TYPE_NUMBER},
			{Name: "quantity", Type: schema.TYPE_NUMBER},
		},
		PKColumns: []int{0, 1}, // composite primary key
	}

	tests := []struct {
		name     string
		colIndex int
		want     bool
	}{
		{name: "order_id is part of PK", colIndex: 0, want: true},
		{name: "item_id is part of PK", colIndex: 1, want: true},
		{name: "quantity is not PK", colIndex: 2, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &table.Columns[tt.colIndex]
			got := isPrimaryKey(col, table)
			if got != tt.want {
				t.Errorf("isPrimaryKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper to create a test table for RowsEventToChanges tests
func makeTestTable() *schema.Table {
	return &schema.Table{
		Name:   "users",
		Schema: "testdb",
		Columns: []schema.TableColumn{
			{Name: "id", Type: schema.TYPE_NUMBER},
			{Name: "name", Type: schema.TYPE_STRING},
			{Name: "email", Type: schema.TYPE_STRING},
		},
		PKColumns: []int{0},
	}
}

func TestRowsEventToChanges_Insert(t *testing.T) {
	table := makeTestTable()

	// canal.RowsEvent uses Table field which is a *schema.Table
	// We need to use canal package types
	event := &canal.RowsEvent{
		Table:  table,
		Action: canal.InsertAction,
		Rows: [][]interface{}{
			{int64(1), "John Doe", "john@example.com"},
			{int64(2), "Jane Doe", "jane@example.com"},
		},
	}
	pos := mysql.Position{Name: "mysql-bin.000001", Pos: 1234}

	changes := RowsEventToChanges(event, pos)

	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	// Check first change
	change1 := changes[0]
	if change1.Type() != "dml" {
		t.Errorf("expected type 'dml', got %s", change1.Type())
	}
	if change1.GetPosition() != "mysql-bin.000001:1234" {
		t.Errorf("expected position 'mysql-bin.000001:1234', got %s", change1.GetPosition())
	}

	dml1, ok := change1.Data.(*types.DMLData)
	if !ok {
		t.Fatal("expected *types.DMLData")
	}
	if dml1.Kind != "insert" {
		t.Errorf("expected kind 'insert', got %s", dml1.Kind)
	}
	if dml1.Table != "users" {
		t.Errorf("expected table 'users', got %s", dml1.Table)
	}
	if len(dml1.ColumnNames) != 3 {
		t.Errorf("expected 3 column names, got %d", len(dml1.ColumnNames))
	}
	if len(dml1.ColumnValues) != 3 {
		t.Errorf("expected 3 column values, got %d", len(dml1.ColumnValues))
	}
	if dml1.OldKeys != nil {
		t.Errorf("expected nil OldKeys for insert")
	}

	// Verify column names
	expectedCols := []string{"id", "name", "email"}
	for i, expected := range expectedCols {
		if dml1.ColumnNames[i] != expected {
			t.Errorf("column %d: expected %s, got %s", i, expected, dml1.ColumnNames[i])
		}
	}

	// Verify first row values
	if dml1.ColumnValues[0].ColumnValue.GetIntValue() != 1 {
		t.Errorf("expected id=1, got %v", dml1.ColumnValues[0].ColumnValue.GetValue())
	}
	if dml1.ColumnValues[1].ColumnValue.GetStringValue() != "John Doe" {
		t.Errorf("expected name='John Doe', got %v", dml1.ColumnValues[1].ColumnValue.GetValue())
	}
}

func TestRowsEventToChanges_Update(t *testing.T) {
	table := makeTestTable()

	// For updates, rows come in pairs: [old, new]
	event := &canal.RowsEvent{
		Table:  table,
		Action: canal.UpdateAction,
		Rows: [][]interface{}{
			{int64(1), "John Doe", "john@example.com"},     // old
			{int64(1), "John Smith", "john@example.com"},   // new
		},
	}
	pos := mysql.Position{Name: "mysql-bin.000001", Pos: 5678}

	changes := RowsEventToChanges(event, pos)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	dml, ok := change.Data.(*types.DMLData)
	if !ok {
		t.Fatal("expected *types.DMLData")
	}

	if dml.Kind != "update" {
		t.Errorf("expected kind 'update', got %s", dml.Kind)
	}

	// Should have OldKeys
	if dml.OldKeys == nil {
		t.Fatal("expected OldKeys for update")
	}
	if len(dml.OldKeys.KeyNames) != 1 {
		t.Errorf("expected 1 key name, got %d", len(dml.OldKeys.KeyNames))
	}
	if dml.OldKeys.KeyNames[0] != "id" {
		t.Errorf("expected key name 'id', got %s", dml.OldKeys.KeyNames[0])
	}
	if dml.OldKeys.KeyValues[0].ColumnValue.GetIntValue() != 1 {
		t.Errorf("expected old key value 1, got %v", dml.OldKeys.KeyValues[0].ColumnValue.GetValue())
	}
}

func TestRowsEventToChanges_Delete(t *testing.T) {
	table := makeTestTable()

	event := &canal.RowsEvent{
		Table:  table,
		Action: canal.DeleteAction,
		Rows: [][]interface{}{
			{int64(42), "Deleted User", "deleted@example.com"},
		},
	}
	pos := mysql.Position{Name: "mysql-bin.000002", Pos: 9999}

	changes := RowsEventToChanges(event, pos)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	dml, ok := change.Data.(*types.DMLData)
	if !ok {
		t.Fatal("expected *types.DMLData")
	}

	if dml.Kind != "delete" {
		t.Errorf("expected kind 'delete', got %s", dml.Kind)
	}

	// Delete should have empty column names/values
	if len(dml.ColumnNames) != 0 {
		t.Errorf("expected 0 column names for delete, got %d", len(dml.ColumnNames))
	}

	// Should have OldKeys
	if dml.OldKeys == nil {
		t.Fatal("expected OldKeys for delete")
	}
	if len(dml.OldKeys.KeyNames) != 1 {
		t.Errorf("expected 1 key name, got %d", len(dml.OldKeys.KeyNames))
	}
	if dml.OldKeys.KeyNames[0] != "id" {
		t.Errorf("expected key name 'id', got %s", dml.OldKeys.KeyNames[0])
	}
	if dml.OldKeys.KeyValues[0].ColumnValue.GetIntValue() != 42 {
		t.Errorf("expected old key value 42, got %v", dml.OldKeys.KeyValues[0].ColumnValue.GetValue())
	}
}

func TestRowsEventToChanges_EmptyRows(t *testing.T) {
	table := makeTestTable()

	event := &canal.RowsEvent{
		Table:  table,
		Action: canal.InsertAction,
		Rows:   [][]interface{}{},
	}
	pos := mysql.Position{Name: "mysql-bin.000001", Pos: 100}

	changes := RowsEventToChanges(event, pos)

	if len(changes) != 0 {
		t.Errorf("expected 0 changes for empty rows, got %d", len(changes))
	}
}

func TestRowsEventToChanges_Update_OddRows(t *testing.T) {
	// Test edge case: odd number of rows in update (should skip incomplete pair)
	table := makeTestTable()

	event := &canal.RowsEvent{
		Table:  table,
		Action: canal.UpdateAction,
		Rows: [][]interface{}{
			{int64(1), "Old", "old@example.com"},
			{int64(1), "New", "new@example.com"},
			{int64(2), "Incomplete", "incomplete@example.com"}, // no pair
		},
	}
	pos := mysql.Position{Name: "mysql-bin.000001", Pos: 100}

	changes := RowsEventToChanges(event, pos)

	// Should only get 1 change (the complete pair)
	if len(changes) != 1 {
		t.Errorf("expected 1 change for odd rows, got %d", len(changes))
	}
}

func TestConvertToProtoChange_MySQL_DMLData(t *testing.T) {
	tests := []struct {
		name   string
		change types.Change
		want   *proto.Change
	}{
		{
			name: "DML insert with MySQL position",
			change: types.Change{
				Position: "mysql-bin.000001:1234",
				Data: &types.DMLData{
					Table:       "testdb.users",
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
				Position: "mysql-bin.000001:1234",
				Type:     "dml",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:       "testdb.users",
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
			name: "DML delete with old keys",
			change: types.Change{
				Position: "binlog.000005:5678",
				Data: &types.DMLData{
					Table:        "testdb.users",
					Kind:         "delete",
					ColumnNames:  []string{},
					ColumnValues: []types.ColumnValueWrapper{},
					OldKeys: &struct {
						KeyNames  []string                   `json:"keynames"`
						KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
					}{
						KeyNames: []string{"id"},
						KeyValues: []types.ColumnValueWrapper{
							{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 42}}},
						},
					},
				},
			},
			want: &proto.Change{
				Position: "binlog.000005:5678",
				Type:     "dml",
				Data: &proto.Change_Dml{
					Dml: &proto.DMLData{
						Table:        "testdb.users",
						Kind:         "delete",
						ColumnNames:  []string{},
						ColumnValues: []*proto.ColumnValue{},
						OldKeys: &proto.OldKeys{
							KeyNames: []string{"id"},
							KeyValues: []*proto.ColumnValue{
								{Value: &proto.ColumnValue_IntValue{IntValue: 42}},
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
			if got.Position != tt.want.Position {
				t.Errorf("Position = %v, want %v", got.Position, tt.want.Position)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
		})
	}
}

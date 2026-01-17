package server

import (
	"testing"
	"time"

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
			position: "0/0",
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
				if got.GetLSN() != "mysql-bin.000001:1234" {
					t.Errorf("QueryEventToChange().GetLSN() = %v, want mysql-bin.000001:1234", got.GetLSN())
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
			name:  "float64 value",
			value: float64(3.14),
			check: func(t *testing.T, cv types.ColumnValueWrapper) {
				if cv.ColumnValue.GetFloatValue() != 3.14 {
					t.Errorf("expected float 3.14, got %v", cv.ColumnValue.GetValue())
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toColumnValue(tt.value, col)
			tt.check(t, got)
		})
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
				LSN: "mysql-bin.000001:1234",
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
				LSN: "binlog.000005:5678",
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

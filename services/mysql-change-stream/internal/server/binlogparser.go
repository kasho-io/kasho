package server

import (
	"fmt"
	"strings"
	"time"

	"kasho/pkg/types"
	"kasho/proto"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-mysql-org/go-mysql/schema"
)

// FormatBinlogPosition converts a MySQL position to our string format
// Format: "mysql-bin.000001:4" (filename:offset)
func FormatBinlogPosition(pos mysql.Position) string {
	return fmt.Sprintf("%s:%d", pos.Name, pos.Pos)
}

// ParseBinlogPosition parses our position string format back to mysql.Position
func ParseBinlogPosition(position string) (mysql.Position, error) {
	if position == "" || position == "bootstrap" {
		return mysql.Position{}, nil
	}

	parts := strings.Split(position, ":")
	if len(parts) != 2 {
		return mysql.Position{}, fmt.Errorf("invalid position format: %s", position)
	}

	var pos uint32
	if _, err := fmt.Sscanf(parts[1], "%d", &pos); err != nil {
		return mysql.Position{}, fmt.Errorf("invalid position offset: %s", parts[1])
	}

	return mysql.Position{
		Name: parts[0],
		Pos:  pos,
	}, nil
}

// RowsEventToChanges converts a canal RowsEvent to our Change types
func RowsEventToChanges(e *canal.RowsEvent, pos mysql.Position) []types.Change {
	var changes []types.Change
	position := FormatBinlogPosition(pos)

	tableName := fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Name)

	switch e.Action {
	case canal.InsertAction:
		for _, row := range e.Rows {
			dml := &types.DMLData{
				Table:        tableName,
				Kind:         "insert",
				ColumnNames:  make([]string, 0, len(e.Table.Columns)),
				ColumnValues: make([]types.ColumnValueWrapper, 0, len(row)),
			}

			for i, col := range e.Table.Columns {
				dml.ColumnNames = append(dml.ColumnNames, col.Name)
				if i < len(row) {
					dml.ColumnValues = append(dml.ColumnValues, toColumnValue(row[i], &col))
				}
			}

			changes = append(changes, types.Change{Position: position, Data: dml})
		}

	case canal.UpdateAction:
		// Rows come in pairs: [old1, new1, old2, new2, ...]
		for i := 0; i < len(e.Rows); i += 2 {
			if i+1 >= len(e.Rows) {
				break
			}
			oldRow := e.Rows[i]
			newRow := e.Rows[i+1]

			dml := &types.DMLData{
				Table:        tableName,
				Kind:         "update",
				ColumnNames:  make([]string, 0),
				ColumnValues: make([]types.ColumnValueWrapper, 0),
			}

			// Find primary key columns
			var pkIndices []int
			for idx, col := range e.Table.Columns {
				if isPrimaryKey(&col, e.Table) {
					pkIndices = append(pkIndices, idx)
				}
			}

			// Build OldKeys from primary key columns
			dml.OldKeys = &struct {
				KeyNames  []string                   `json:"keynames"`
				KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
			}{
				KeyNames:  make([]string, 0, len(pkIndices)),
				KeyValues: make([]types.ColumnValueWrapper, 0, len(pkIndices)),
			}

			for _, idx := range pkIndices {
				col := e.Table.Columns[idx]
				dml.OldKeys.KeyNames = append(dml.OldKeys.KeyNames, col.Name)
				if idx < len(oldRow) {
					dml.OldKeys.KeyValues = append(dml.OldKeys.KeyValues, toColumnValue(oldRow[idx], &col))
				}
			}

			// Include columns that changed or all non-PK columns
			for idx, col := range e.Table.Columns {
				if idx < len(newRow) && idx < len(oldRow) {
					// Include if value changed or if it's not a PK column
					if !isPrimaryKey(&col, e.Table) || oldRow[idx] != newRow[idx] {
						dml.ColumnNames = append(dml.ColumnNames, col.Name)
						dml.ColumnValues = append(dml.ColumnValues, toColumnValue(newRow[idx], &col))
					}
				}
			}

			changes = append(changes, types.Change{Position: position, Data: dml})
		}

	case canal.DeleteAction:
		for _, row := range e.Rows {
			dml := &types.DMLData{
				Table:        tableName,
				Kind:         "delete",
				ColumnNames:  make([]string, 0),
				ColumnValues: make([]types.ColumnValueWrapper, 0),
			}

			// Build OldKeys from primary key columns
			var pkIndices []int
			for idx, col := range e.Table.Columns {
				if isPrimaryKey(&col, e.Table) {
					pkIndices = append(pkIndices, idx)
				}
			}

			dml.OldKeys = &struct {
				KeyNames  []string                   `json:"keynames"`
				KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
			}{
				KeyNames:  make([]string, 0, len(pkIndices)),
				KeyValues: make([]types.ColumnValueWrapper, 0, len(pkIndices)),
			}

			for _, idx := range pkIndices {
				col := e.Table.Columns[idx]
				dml.OldKeys.KeyNames = append(dml.OldKeys.KeyNames, col.Name)
				if idx < len(row) {
					dml.OldKeys.KeyValues = append(dml.OldKeys.KeyValues, toColumnValue(row[idx], &col))
				}
			}

			changes = append(changes, types.Change{Position: position, Data: dml})
		}
	}

	return changes
}

// QueryEventToChange converts a DDL query event to a Change
func QueryEventToChange(header *replication.EventHeader, e *replication.QueryEvent, pos mysql.Position) *types.Change {
	query := string(e.Query)

	// Skip non-DDL queries
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(upperQuery, "CREATE") &&
		!strings.HasPrefix(upperQuery, "ALTER") &&
		!strings.HasPrefix(upperQuery, "DROP") &&
		!strings.HasPrefix(upperQuery, "RENAME") &&
		!strings.HasPrefix(upperQuery, "TRUNCATE") {
		return nil
	}

	position := FormatBinlogPosition(pos)

	// Use the event header timestamp (Unix timestamp when event occurred on the server)
	eventTime := time.Unix(int64(header.Timestamp), 0)

	ddl := types.DDLData{
		ID:       0, // MySQL doesn't have a DDL ID like PostgreSQL
		Time:     eventTime,
		Username: "", // Not available from binlog
		Database: string(e.Schema),
		DDL:      query,
	}

	return &types.Change{Position: position, Data: ddl}
}

// isPrimaryKey checks if a column is part of the primary key
func isPrimaryKey(col *schema.TableColumn, table *schema.Table) bool {
	for _, pkIdx := range table.PKColumns {
		if pkIdx < len(table.Columns) && table.Columns[pkIdx].Name == col.Name {
			return true
		}
	}
	return false
}

// toColumnValue converts a MySQL value to our ColumnValueWrapper
func toColumnValue(value any, col *schema.TableColumn) types.ColumnValueWrapper {
	if value == nil {
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: ""}}}
	}

	switch v := value.(type) {
	case string:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: v}}}
	case []byte:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: string(v)}}}
	case int8:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case int16:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case int32:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case int64:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: v}}}
	case int:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case uint8:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case uint16:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case uint32:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case uint64:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case uint:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case float32:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: float64(v)}}}
	case float64:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: v}}}
	case bool:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: v}}}
	case time.Time:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: v.Format(time.RFC3339)}}}
	default:
		// For any other type, convert to string
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: fmt.Sprint(v)}}}
	}
}

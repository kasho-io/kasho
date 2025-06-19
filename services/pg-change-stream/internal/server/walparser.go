package server

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"kasho/pkg/types"
	"kasho/proto"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgproto3"
)

var relationMap = make(map[uint32]*pglogrepl.RelationMessageV2)

func ParseMessage(msg pgproto3.BackendMessage) ([]types.Change, pglogrepl.LSN, error) {
	copyData, ok := msg.(*pgproto3.CopyData)
	if !ok {
		return nil, 0, nil
	}

	if copyData.Data[0] != pglogrepl.XLogDataByteID {
		return nil, 0, nil
	}

	walData, err := pglogrepl.ParseXLogData(copyData.Data[1:])
	if err != nil {
		return nil, 0, fmt.Errorf("error parsing WAL data: %w", err)
	}

	changes, err := ParseWALData(walData.WALData, walData.WALStart)
	if err != nil {
		return nil, 0, err
	}

	return changes, walData.WALStart, nil
}

func decodeColumnData(col *pglogrepl.TupleDataColumn, colType uint32) (any, error) {
	if col == nil {
		return nil, nil
	}

	// For text format, we need to parse the string value based on the column type
	strValue := string(col.Data)
	if strValue == "" {
		return nil, nil
	}

	switch colType {
	case 21: // int2
		val, err := strconv.ParseInt(strValue, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid int2 value: %s", strValue)
		}
		return int16(val), nil
	case 23: // int4
		val, err := strconv.ParseInt(strValue, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid int4 value: %s", strValue)
		}
		return int32(val), nil
	case 20: // int8
		val, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int8 value: %s", strValue)
		}
		return val, nil
	case 700: // float4
		val, err := strconv.ParseFloat(strValue, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid float4 value: %s", strValue)
		}
		return float32(val), nil
	case 701: // float8
		val, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float8 value: %s", strValue)
		}
		return val, nil
	case 16: // bool
		switch strValue {
		case "t", "true", "1":
			return true, nil
		case "f", "false", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool value: %s", strValue)
		}
	case 1114: // timestamp
		t, err := time.Parse("2006-01-02 15:04:05.999999", strValue)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp value: %s", strValue)
		}
		return t, nil
	case 1184: // timestamptz
		t, err := time.Parse("2006-01-02 15:04:05.999999-07", strValue)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamptz value: %s", strValue)
		}
		return t, nil
	case 1082: // date
		t, err := time.Parse("2006-01-02", strValue)
		if err != nil {
			return nil, fmt.Errorf("invalid date value: %s", strValue)
		}
		return t, nil
	case 25, 1043: // text, varchar
		return strValue, nil
	default:
		return strValue, nil
	}
}

func toColumnValue(value any) types.ColumnValueWrapper {
	switch v := value.(type) {
	case string:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: v}}}
	case int32:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(v)}}}
	case int64:
		return types.ColumnValueWrapper{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: v}}}
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

func ParseWALData(walData []byte, lsn pglogrepl.LSN) ([]types.Change, error) {
	msg, err := pglogrepl.ParseV2(walData, false)
	if err != nil {
		return nil, fmt.Errorf("error parsing WAL message: %w", err)
	}

	var changes []types.Change

	switch v := msg.(type) {
	case *pglogrepl.RelationMessageV2:
		relationMap[v.RelationID] = v

	case *pglogrepl.InsertMessageV2:
		rel, ok := relationMap[v.RelationID]
		if !ok {
			return nil, fmt.Errorf("unknown relation ID %d", v.RelationID)
		}

		tableName := fmt.Sprintf("%s.%s", rel.Namespace, rel.RelationName)
		if tableName == "public.kasho_ddl_log" {
			ddl := types.DDLData{}
			for i, col := range rel.Columns {
				if i < len(v.Tuple.Columns) {
					colData := v.Tuple.Columns[i]
					if colData == nil {
						continue
					}
					value, err := decodeColumnData(colData, col.DataType)
					if err != nil {
						return nil, fmt.Errorf("error decoding column %s: %w", col.Name, err)
					}
					switch col.Name {
					case "id":
						if id, ok := value.(int32); ok {
							ddl.ID = int(id)
						}
					case "time":
						if t, ok := value.(time.Time); ok {
							ddl.Time = t
						}
					case "username":
						if username, ok := value.(string); ok {
							ddl.Username = username
						}
					case "database":
						if db, ok := value.(string); ok {
							ddl.Database = db
						}
					case "ddl":
						if ddlStr, ok := value.(string); ok {
							ddl.DDL = ddlStr
						}
					}
				}
			}
			changes = append(changes, types.Change{LSN: lsn.String(), Data: ddl})
		} else {
			dml := types.DMLData{
				Table:        tableName,
				Kind:         "insert",
				ColumnNames:  make([]string, 0, len(rel.Columns)),
				ColumnValues: make([]types.ColumnValueWrapper, 0, len(v.Tuple.Columns)),
			}

			for i, col := range rel.Columns {
				dml.ColumnNames = append(dml.ColumnNames, col.Name)
				if i < len(v.Tuple.Columns) {
					colData := v.Tuple.Columns[i]
					if colData == nil {
						continue
					}
					value, err := decodeColumnData(colData, col.DataType)
					if err != nil {
						return nil, fmt.Errorf("error decoding column %s: %w", col.Name, err)
					}
					dml.ColumnValues = append(dml.ColumnValues, toColumnValue(value))
				}
			}

			changes = append(changes, types.Change{LSN: lsn.String(), Data: dml})
		}

	case *pglogrepl.UpdateMessageV2:
		rel, ok := relationMap[v.RelationID]
		if !ok {
			return nil, fmt.Errorf("unknown relation ID %d", v.RelationID)
		}

		// Find primary key columns
		var pkColumns []*pglogrepl.RelationMessageColumn
		for _, col := range rel.Columns {
			if col.Flags == 1 { // 1 indicates the column is part of the key
				pkColumns = append(pkColumns, col)
			}
		}

		dml := types.DMLData{
			Table:        fmt.Sprintf("%s.%s", rel.Namespace, rel.RelationName),
			Kind:         "update",
			ColumnNames:  make([]string, 0),
			ColumnValues: make([]types.ColumnValueWrapper, 0),
		}

		// Initialize OldKeys with primary key columns
		dml.OldKeys = &struct {
			KeyNames  []string                   `json:"keynames"`
			KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
		}{
			KeyNames:  make([]string, 0, len(pkColumns)),
			KeyValues: make([]types.ColumnValueWrapper, 0, len(pkColumns)),
		}

		// Store old values for comparison
		// NOTE: What OldTuple contains depends on the table's REPLICA IDENTITY setting:
		// - DEFAULT: Only primary key columns
		// - FULL: All columns with their old values
		// - NOTHING: Empty (no old values)
		// - USING INDEX: Only columns in the specified index
		oldValues := make(map[string]any)
		if v.OldTuple != nil && len(v.OldTuple.Columns) > 0 {
			for i, col := range rel.Columns {
				if i < len(v.OldTuple.Columns) {
					colData := v.OldTuple.Columns[i]
					if colData != nil {
						value, err := decodeColumnData(colData, col.DataType)
						if err != nil {
							return nil, fmt.Errorf("error decoding old column %s: %w", col.Name, err)
						}
						oldValues[col.Name] = value
					}
				}
			}
		}

		// Process NewTuple which ALWAYS contains ALL columns regardless of REPLICA IDENTITY
		//
		// This code compares values from NewTuple against oldValues to determine what to include.
		// With REPLICA IDENTITY DEFAULT, oldValues only contains primary key columns.
		// As a result:
		// - Primary key columns: Added to OldKeys, only included in output if changed
		// - Non-PK columns: Not found in oldValues, so all are included in output
		//
		// This will cause UPDATE statements to include all columns even when only some were changed.
		for i, col := range rel.Columns {
			if i < len(v.NewTuple.Columns) {
				colData := v.NewTuple.Columns[i]
				if colData == nil {
					continue
				}
				newValue, err := decodeColumnData(colData, col.DataType)
				if err != nil {
					return nil, fmt.Errorf("error decoding column %s: %w", col.Name, err)
				}

				// If this is a primary key, add to OldKeys
				if col.Flags == 1 {
					dml.OldKeys.KeyNames = append(dml.OldKeys.KeyNames, col.Name)
					dml.OldKeys.KeyValues = append(dml.OldKeys.KeyValues, toColumnValue(newValue))
				}

				// Check if the value has changed
				oldValue, exists := oldValues[col.Name]
				if !exists {
					// Column not in oldValues - with REPLICA IDENTITY DEFAULT, this includes
					// all non-PK columns, so they all get added to the output
					if v.OldTuple == nil && col.Flags != 1 {
						dml.ColumnNames = append(dml.ColumnNames, col.Name)
						dml.ColumnValues = append(dml.ColumnValues, toColumnValue(newValue))
					}
				} else if oldValue != newValue {
					// Value has actually changed
					dml.ColumnNames = append(dml.ColumnNames, col.Name)
					dml.ColumnValues = append(dml.ColumnValues, toColumnValue(newValue))
				}
			}
		}

		changes = append(changes, types.Change{LSN: lsn.String(), Data: dml})

	case *pglogrepl.DeleteMessageV2:
		rel, ok := relationMap[v.RelationID]
		if !ok {
			return nil, fmt.Errorf("unknown relation ID %d", v.RelationID)
		}

		dml := types.DMLData{
			Table:        fmt.Sprintf("%s.%s", rel.Namespace, rel.RelationName),
			Kind:         "delete",
			ColumnNames:  make([]string, 0, len(rel.Columns)),
			ColumnValues: make([]types.ColumnValueWrapper, 0, len(v.OldTuple.Columns)),
		}

		// Add old key values
		dml.OldKeys = &struct {
			KeyNames  []string                   `json:"keynames"`
			KeyValues []types.ColumnValueWrapper `json:"keyvalues"`
		}{
			KeyNames:  make([]string, 0, len(rel.Columns)),
			KeyValues: make([]types.ColumnValueWrapper, 0, len(v.OldTuple.Columns)),
		}

		for i, col := range rel.Columns {
			if i < len(v.OldTuple.Columns) {
				colData := v.OldTuple.Columns[i]
				if colData == nil {
					continue
				}
				dml.OldKeys.KeyNames = append(dml.OldKeys.KeyNames, col.Name)
				value, err := decodeColumnData(colData, col.DataType)
				if err != nil {
					return nil, fmt.Errorf("error decoding old key column %s: %w", col.Name, err)
				}
				dml.OldKeys.KeyValues = append(dml.OldKeys.KeyValues, toColumnValue(value))
			}
		}

		changes = append(changes, types.Change{LSN: lsn.String(), Data: dml})

	case *pglogrepl.BeginMessage:
		// No changes to create for begin messages

	case *pglogrepl.CommitMessage:
		// No changes to create for commit messages

	default:
		log.Printf("Unhandled message type: %T", msg)
	}

	return changes, nil
}

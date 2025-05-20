package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"pg-change-stream/internal/types"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgproto3"
)

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

func ParseWALData(walData []byte, lsn pglogrepl.LSN) ([]types.Change, error) {
	jsonStart := bytes.Index(walData, []byte("{"))
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON data found in WAL")
	}
	jsonEnd := bytes.LastIndex(walData, []byte("}"))
	if jsonEnd == -1 {
		return nil, fmt.Errorf("invalid JSON data in WAL")
	}

	jsonData := walData[jsonStart : jsonEnd+1]
	var event map[string]any
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return nil, fmt.Errorf("error parsing WAL JSON: %w", err)
	}

	changes, ok := event["change"].([]any)
	if !ok {
		return nil, fmt.Errorf("no changes found in WAL data")
	}

	var result []types.Change
	for _, c := range changes {
		change, ok := c.(map[string]any)
		if !ok {
			continue
		}

		table, ok := change["table"].(string)
		if !ok {
			continue
		}

		if table == "translicate_ddl_log" && change["kind"].(string) == "insert" {
			ddl := types.DDLData{}
			for i, col := range change["columnnames"].([]any) {
				colName := col.(string)
				value := change["columnvalues"].([]any)[i]
				switch colName {
				case "id":
					if value != nil {
						ddl.ID = int(value.(float64))
					}
				case "time":
					if value != nil {
						ddl.Time = value.(time.Time)
					}
				case "username":
					if value != nil {
						ddl.Username = value.(string)
					}
				case "database":
					if value != nil {
						ddl.Database = value.(string)
					}
				case "ddl":
					if value != nil {
						ddl.DDL = value.(string)
					}
				}
			}
			result = append(result, types.Change{LSN: lsn.String(), Data: ddl})
		} else {
			dml := types.DMLData{
				Table:        table,
				ColumnNames:  make([]string, 0),
				ColumnValues: make([]any, 0),
				Kind:         change["kind"].(string),
			}

			if names, ok := change["columnnames"].([]any); ok {
				for _, n := range names {
					dml.ColumnNames = append(dml.ColumnNames, n.(string))
				}
			}
			if values, ok := change["columnvalues"].([]any); ok {
				dml.ColumnValues = values
			}

			if oldKeys, ok := change["oldkeys"].(map[string]any); ok {
				dml.OldKeys = &struct {
					KeyNames  []string `json:"keynames"`
					KeyValues []any    `json:"keyvalues"`
				}{
					KeyNames:  make([]string, 0),
					KeyValues: make([]any, 0),
				}
				if names, ok := oldKeys["keynames"].([]any); ok {
					for _, n := range names {
						dml.OldKeys.KeyNames = append(dml.OldKeys.KeyNames, n.(string))
					}
				}
				if values, ok := oldKeys["keyvalues"].([]any); ok {
					dml.OldKeys.KeyValues = values
				}
			}

			result = append(result, types.Change{LSN: lsn.String(), Data: dml})
		}
	}

	return result, nil
}

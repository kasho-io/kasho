package converter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"kasho/pkg/types"
	"kasho/proto"
	"mysql-bootstrap-sync/internal/parser"
)

// ChangeConverter converts parsed statements to Change objects
type ChangeConverter struct {
	positionGenerator *PositionGenerator
}

// NewChangeConverter creates a new change converter
func NewChangeConverter() *ChangeConverter {
	posGen := NewPositionGenerator()

	return &ChangeConverter{
		positionGenerator: posGen,
	}
}

// ConvertStatements converts a list of parsed statements to Change objects
func (c *ChangeConverter) ConvertStatements(statements []parser.Statement) ([]*types.Change, error) {
	changes := make([]*types.Change, 0, len(statements))

	for _, stmt := range statements {
		switch s := stmt.(type) {
		case parser.DDLStatement:
			change, err := c.convertDDLStatement(s)
			if err != nil {
				return nil, fmt.Errorf("failed to convert DDL statement: %w", err)
			}
			changes = append(changes, change)

		case parser.DMLStatement:
			// DML statements might generate multiple changes (one per row)
			dmlChanges, err := c.convertDMLStatement(s)
			if err != nil {
				return nil, fmt.Errorf("failed to convert DML statement: %w", err)
			}
			changes = append(changes, dmlChanges...)

		default:
			// Skip unknown statement types
			continue
		}
	}

	return changes, nil
}

// convertDDLStatement converts a DDL statement to a Change object
func (c *ChangeConverter) convertDDLStatement(stmt parser.DDLStatement) (*types.Change, error) {
	pos := c.positionGenerator.Next()

	ddlData := &types.DDLData{
		ID:       int(c.positionGenerator.GetSequence()), // Use sequence as ID
		Time:     stmt.Time,
		Username: "bootstrap", // Bootstrap user
		Database: stmt.Database,
		DDL:      stmt.SQL,
	}

	change := &types.Change{
		LSN:  pos,
		Data: ddlData,
	}

	return change, nil
}

// convertDMLStatement converts a DML statement to Change objects (one per row)
func (c *ChangeConverter) convertDMLStatement(stmt parser.DMLStatement) ([]*types.Change, error) {
	changes := make([]*types.Change, 0, len(stmt.ColumnValues))

	for _, row := range stmt.ColumnValues {
		// Allow empty column names (INSERT without explicit columns)
		if len(stmt.ColumnNames) > 0 && len(row) != len(stmt.ColumnNames) {
			return nil, fmt.Errorf("column count mismatch: expected %d, got %d", len(stmt.ColumnNames), len(row))
		}

		pos := c.positionGenerator.Next()

		// Convert row values to ColumnValueWrapper objects
		columnValues := make([]types.ColumnValueWrapper, len(row))
		for i, value := range row {
			columnValue, err := c.convertValue(value)
			if err != nil {
				var colName string
				if i < len(stmt.ColumnNames) {
					colName = stmt.ColumnNames[i]
				} else {
					colName = fmt.Sprintf("column_%d", i)
				}
				return nil, fmt.Errorf("failed to convert value for column %s: %w", colName, err)
			}
			columnValues[i] = types.ColumnValueWrapper{ColumnValue: columnValue}
		}

		dmlData := &types.DMLData{
			Table:        stmt.Table,
			ColumnNames:  stmt.ColumnNames,
			ColumnValues: columnValues,
			Kind:         "insert", // All bootstrap data is inserts
			OldKeys:      nil,      // No old keys for inserts
		}

		change := &types.Change{
			LSN:  pos,
			Data: dmlData,
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// convertValue converts a string value to a ColumnValue protobuf object
func (c *ChangeConverter) convertValue(value string) (*proto.ColumnValue, error) {
	// Handle NULL values (empty strings from our parser)
	if value == "" {
		// Return a ColumnValue with no value set (representing NULL)
		return &proto.ColumnValue{}, nil
	}

	// Try to detect the type of the value and convert accordingly

	// Try integer
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &proto.ColumnValue{
			Value: &proto.ColumnValue_IntValue{IntValue: intValue},
		}, nil
	}

	// Try float
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return &proto.ColumnValue{
			Value: &proto.ColumnValue_FloatValue{FloatValue: floatValue},
		}, nil
	}

	// Try boolean
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return &proto.ColumnValue{
			Value: &proto.ColumnValue_BoolValue{BoolValue: boolValue},
		}, nil
	}

	// Try timestamp (ISO 8601 format and MySQL formats)
	if strings.Contains(value, "T") || strings.Contains(value, "-") {
		if _, err := time.Parse(time.RFC3339, value); err == nil {
			return &proto.ColumnValue{
				Value: &proto.ColumnValue_TimestampValue{TimestampValue: value},
			}, nil
		}
		// Try other common timestamp formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.000000",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return &proto.ColumnValue{
					Value: &proto.ColumnValue_TimestampValue{TimestampValue: t.Format(time.RFC3339)},
				}, nil
			}
		}
	}

	// Default to string value
	return &proto.ColumnValue{
		Value: &proto.ColumnValue_StringValue{StringValue: value},
	}, nil
}

// GetNextPosition returns the next position that would be generated
func (c *ChangeConverter) GetNextPosition() string {
	return c.positionGenerator.Peek()
}

// GetCurrentSequence returns the current sequence number
func (c *ChangeConverter) GetCurrentSequence() int64 {
	return c.positionGenerator.GetSequence()
}

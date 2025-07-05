package converter

import (
	"testing"

	"kasho/pkg/types"
	"kasho/proto"
	"pg-bootstrap-sync/internal/parser"
)

func TestChangeConverter_ConvertDDLStatement(t *testing.T) {
	converter := NewChangeConverter()
	
	ddlStmt := parser.DDLStatement{
		SQL:      "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100));",
		Table:    "users",
		Database: "testdb",
	}
	
	change, err := converter.convertDDLStatement(ddlStmt)
	if err != nil {
		t.Fatalf("convertDDLStatement failed: %v", err)
	}
	
	// Verify change structure
	if change.Type() != "ddl" {
		t.Errorf("Expected type 'ddl', got %q", change.Type())
	}
	
	if change.LSN != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("Expected LSN '0/BOOTSTRAP0000000000000001', got %q", change.LSN)
	}
	
	ddlData, ok := change.Data.(*types.DDLData)
	if !ok {
		t.Fatal("Change data is not DDLData")
	}
	
	if ddlData.DDL != ddlStmt.SQL {
		t.Errorf("Expected DDL %q, got %q", ddlStmt.SQL, ddlData.DDL)
	}
	
	if ddlData.Username != "bootstrap" {
		t.Errorf("Expected username 'bootstrap', got %q", ddlData.Username)
	}
}

func TestChangeConverter_ConvertDMLStatement(t *testing.T) {
	converter := NewChangeConverter()
	
	dmlStmt := parser.DMLStatement{
		Table:       "users",
		ColumnNames: []string{"id", "name", "email"},
		ColumnValues: [][]string{
			{"1", "John Doe", "john@example.com"},
			{"2", "Jane Smith", "jane@example.com"},
		},
	}
	
	changes, err := converter.convertDMLStatement(dmlStmt)
	if err != nil {
		t.Fatalf("convertDMLStatement failed: %v", err)
	}
	
	// Should generate one change per row
	if len(changes) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(changes))
	}
	
	// Verify first change
	change1 := changes[0]
	if change1.Type() != "dml" {
		t.Errorf("Expected type 'dml', got %q", change1.Type())
	}
	
	if change1.LSN != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("Expected LSN '0/BOOTSTRAP0000000000000001', got %q", change1.LSN)
	}
	
	dmlData1, ok := change1.Data.(*types.DMLData)
	if !ok {
		t.Fatal("Change data is not DMLData")
	}
	
	if dmlData1.Table != "users" {
		t.Errorf("Expected table 'users', got %q", dmlData1.Table)
	}
	
	if dmlData1.Kind != "insert" {
		t.Errorf("Expected kind 'insert', got %q", dmlData1.Kind)
	}
	
	if len(dmlData1.ColumnNames) != 3 {
		t.Errorf("Expected 3 column names, got %d", len(dmlData1.ColumnNames))
	}
	
	if len(dmlData1.ColumnValues) != 3 {
		t.Errorf("Expected 3 column values, got %d", len(dmlData1.ColumnValues))
	}
	
	// Verify second change has incremented LSN
	change2 := changes[1]
	if change2.LSN != "0/BOOTSTRAP0000000000000002" {
		t.Errorf("Expected LSN '0/BOOTSTRAP0000000000000002', got %q", change2.LSN)
	}
}

func TestChangeConverter_ConvertValue(t *testing.T) {
	converter := NewChangeConverter()
	
	tests := []struct {
		input    string
		expected *proto.ColumnValue
	}{
		{
			input: "",
			expected: &proto.ColumnValue{}, // NULL value
		},
		{
			input: "123",
			expected: &proto.ColumnValue{
				Value: &proto.ColumnValue_IntValue{IntValue: 123},
			},
		},
		{
			input: "123.45",
			expected: &proto.ColumnValue{
				Value: &proto.ColumnValue_FloatValue{FloatValue: 123.45},
			},
		},
		{
			input: "true",
			expected: &proto.ColumnValue{
				Value: &proto.ColumnValue_BoolValue{BoolValue: true},
			},
		},
		{
			input: "hello world",
			expected: &proto.ColumnValue{
				Value: &proto.ColumnValue_StringValue{StringValue: "hello world"},
			},
		},
	}
	
	for _, tt := range tests {
		result, err := converter.convertValue(tt.input)
		if err != nil {
			t.Errorf("convertValue(%q) failed: %v", tt.input, err)
			continue
		}
		
		// Compare the actual values based on type
		switch expectedVal := tt.expected.Value.(type) {
		case nil:
			if result.Value != nil {
				t.Errorf("convertValue(%q): expected NULL, got %T", tt.input, result.Value)
			}
		case *proto.ColumnValue_IntValue:
			if intVal := result.GetIntValue(); intVal != expectedVal.IntValue {
				t.Errorf("convertValue(%q): expected int %d, got %d", tt.input, expectedVal.IntValue, intVal)
			}
		case *proto.ColumnValue_FloatValue:
			if floatVal := result.GetFloatValue(); floatVal != expectedVal.FloatValue {
				t.Errorf("convertValue(%q): expected float %f, got %f", tt.input, expectedVal.FloatValue, floatVal)
			}
		case *proto.ColumnValue_BoolValue:
			if boolVal := result.GetBoolValue(); boolVal != expectedVal.BoolValue {
				t.Errorf("convertValue(%q): expected bool %t, got %t", tt.input, expectedVal.BoolValue, boolVal)
			}
		case *proto.ColumnValue_StringValue:
			if stringVal := result.GetStringValue(); stringVal != expectedVal.StringValue {
				t.Errorf("convertValue(%q): expected string %q, got %q", tt.input, expectedVal.StringValue, stringVal)
			}
		}
	}
}

func TestChangeConverter_ConvertStatements(t *testing.T) {
	converter := NewChangeConverter()
	
	statements := []parser.Statement{
		parser.DDLStatement{
			SQL:   "CREATE TABLE users (id INT);",
			Table: "users",
		},
		parser.DMLStatement{
			Table:        "users",
			ColumnNames:  []string{"id"},
			ColumnValues: [][]string{{"1"}, {"2"}},
		},
	}
	
	changes, err := converter.ConvertStatements(statements)
	if err != nil {
		t.Fatalf("ConvertStatements failed: %v", err)
	}
	
	// Should have 1 DDL + 2 DML = 3 total changes
	if len(changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(changes))
	}
	
	// First change should be DDL
	if changes[0].Type() != "ddl" {
		t.Errorf("First change should be DDL, got %q", changes[0].Type())
	}
	
	// Remaining changes should be DML
	for i := 1; i < len(changes); i++ {
		if changes[i].Type() != "dml" {
			t.Errorf("Change %d should be DML, got %q", i, changes[i].Type())
		}
	}
	
	// Verify LSN sequence
	expectedLSNs := []string{
		"0/BOOTSTRAP0000000000000001",
		"0/BOOTSTRAP0000000000000002", 
		"0/BOOTSTRAP0000000000000003",
	}
	
	for i, expected := range expectedLSNs {
		if changes[i].LSN != expected {
			t.Errorf("Change %d: expected LSN %q, got %q", i, expected, changes[i].LSN)
		}
	}
}
package parser

import (
	"fmt"
	"regexp"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// SQLParser provides advanced SQL parsing using pg_query_go
type SQLParser struct{}

// NewSQLParser creates a new SQL parser
func NewSQLParser() *SQLParser {
	return &SQLParser{}
}

// ParseSQL parses a SQL statement and returns structured information
func (p *SQLParser) ParseSQL(sql string) (*ParsedSQL, error) {
	result, err := pg_query.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	parsed := &ParsedSQL{
		OriginalSQL: sql,
		Statements:  make([]ParsedStatement, 0),
	}

	// Process each statement in the parse tree
	for _, stmt := range result.Stmts {
		parsedStmt, err := p.parseStatement(stmt.Stmt)
		if err != nil {
			// Log error but continue with other statements
			continue
		}
		if parsedStmt != nil {
			parsed.Statements = append(parsed.Statements, *parsedStmt)
		}
	}

	return parsed, nil
}

// ParsedSQL represents the result of parsing SQL with pg_query_go
type ParsedSQL struct {
	OriginalSQL string
	Statements  []ParsedStatement
}

// ParsedStatement represents a single parsed SQL statement
type ParsedStatement struct {
	Type      string                 // Statement type (CREATE_TABLE, INSERT, etc.)
	TableName string                 // Primary table affected
	Columns   []string               // Column names involved
	Values    [][]interface{}        // Values for INSERT statements
	Metadata  map[string]interface{} // Additional metadata
}

// parseStatement parses a single statement node
func (p *SQLParser) parseStatement(stmt *pg_query.Node) (*ParsedStatement, error) {
	if stmt == nil {
		return nil, nil
	}

	switch node := stmt.Node.(type) {
	case *pg_query.Node_CreateStmt:
		return p.parseCreateTable(node.CreateStmt)
	case *pg_query.Node_IndexStmt:
		return p.parseCreateIndex(node.IndexStmt)
	case *pg_query.Node_AlterTableStmt:
		return p.parseAlterTable(node.AlterTableStmt)
	case *pg_query.Node_InsertStmt:
		return p.parseInsert(node.InsertStmt)
	case *pg_query.Node_UpdateStmt:
		return &ParsedStatement{
			Type:      "UPDATE",
			TableName: node.UpdateStmt.Relation.Relname,
		}, nil
	case *pg_query.Node_DeleteStmt:
		return &ParsedStatement{
			Type:      "DELETE",
			TableName: node.DeleteStmt.Relation.Relname,
		}, nil
	case *pg_query.Node_CreateSeqStmt:
		return &ParsedStatement{
			Type:      "CREATE_SEQUENCE",
			TableName: node.CreateSeqStmt.Sequence.Relname,
		}, nil
	case *pg_query.Node_AlterSeqStmt:
		return &ParsedStatement{
			Type:      "ALTER_SEQUENCE",
			TableName: node.AlterSeqStmt.Sequence.Relname,
		}, nil
	case *pg_query.Node_SelectStmt:
		return p.parseSelect(node.SelectStmt)
	case *pg_query.Node_CreateFunctionStmt:
		return &ParsedStatement{
			Type: "CREATE_FUNCTION",
		}, nil
	case *pg_query.Node_CreateTrigStmt:
		return &ParsedStatement{
			Type:      "CREATE_TRIGGER",
			TableName: node.CreateTrigStmt.Relation.Relname,
		}, nil
	case *pg_query.Node_CreateEventTrigStmt:
		return &ParsedStatement{
			Type: "CREATE_EVENT_TRIGGER",
		}, nil
	case *pg_query.Node_DropStmt:
		return &ParsedStatement{
			Type: "DROP",
		}, nil
	case *pg_query.Node_TruncateStmt:
		return &ParsedStatement{
			Type: "TRUNCATE",
		}, nil
	case *pg_query.Node_CommentStmt:
		return &ParsedStatement{
			Type: "COMMENT",
		}, nil
	case *pg_query.Node_GrantStmt:
		return &ParsedStatement{
			Type: "GRANT",
		}, nil
	case *pg_query.Node_VariableSetStmt:
		return &ParsedStatement{
			Type: "SET",
		}, nil
	case *pg_query.Node_TransactionStmt:
		return &ParsedStatement{
			Type: "TRANSACTION",
		}, nil
	default:
		// For unsupported statement types, return basic info
		return &ParsedStatement{
			Type:     "UNKNOWN",
			Metadata: map[string]interface{}{"node_type": fmt.Sprintf("%T", node)},
		}, nil
	}
}

// parseCreateTable parses a CREATE TABLE statement
func (p *SQLParser) parseCreateTable(stmt *pg_query.CreateStmt) (*ParsedStatement, error) {
	if stmt.Relation == nil {
		return nil, fmt.Errorf("CREATE TABLE statement missing relation")
	}

	// Build qualified table name with schema if present
	tableName := stmt.Relation.Relname
	if stmt.Relation.Schemaname != "" {
		tableName = stmt.Relation.Schemaname + "." + tableName
	}
	columns := make([]string, 0)

	// Extract column names
	for _, element := range stmt.TableElts {
		if colDef := element.GetColumnDef(); colDef != nil {
			columns = append(columns, colDef.Colname)
		}
	}

	return &ParsedStatement{
		Type:      "CREATE_TABLE",
		TableName: tableName,
		Columns:   columns,
		Metadata: map[string]interface{}{
			"if_not_exists": stmt.IfNotExists,
			"temporary":     stmt.Relation.Relpersistence == "t",
		},
	}, nil
}

// parseCreateIndex parses a CREATE INDEX statement
func (p *SQLParser) parseCreateIndex(stmt *pg_query.IndexStmt) (*ParsedStatement, error) {
	if stmt.Relation == nil {
		return nil, fmt.Errorf("CREATE INDEX statement missing relation")
	}

	// Build qualified table name with schema if present
	tableName := stmt.Relation.Relname
	if stmt.Relation.Schemaname != "" {
		tableName = stmt.Relation.Schemaname + "." + tableName
	}
	indexName := ""
	if stmt.Idxname != "" {
		indexName = stmt.Idxname
	}

	columns := make([]string, 0)
	// Note: IndexParams parsing is complex in pg_query_go
	// For now, we'll skip detailed column extraction from indexes
	// This can be enhanced later if needed

	return &ParsedStatement{
		Type:      "CREATE_INDEX",
		TableName: tableName,
		Columns:   columns,
		Metadata: map[string]interface{}{
			"index_name": indexName,
			"unique":     stmt.Unique,
		},
	}, nil
}

// parseAlterTable parses an ALTER TABLE statement
func (p *SQLParser) parseAlterTable(stmt *pg_query.AlterTableStmt) (*ParsedStatement, error) {
	if stmt.Relation == nil {
		return nil, fmt.Errorf("ALTER TABLE statement missing relation")
	}

	// Build qualified table name with schema if present
	tableName := stmt.Relation.Relname
	if stmt.Relation.Schemaname != "" {
		tableName = stmt.Relation.Schemaname + "." + tableName
	}
	alterType := "UNKNOWN"

	// Extract the type of alteration
	if len(stmt.Cmds) > 0 {
		if cmd := stmt.Cmds[0].GetAlterTableCmd(); cmd != nil {
			alterType = cmd.Subtype.String()
		}
	}

	return &ParsedStatement{
		Type:      "ALTER_TABLE",
		TableName: tableName,
		Metadata: map[string]interface{}{
			"alter_type": alterType,
		},
	}, nil
}

// parseInsert parses an INSERT statement
func (p *SQLParser) parseInsert(stmt *pg_query.InsertStmt) (*ParsedStatement, error) {
	if stmt.Relation == nil {
		return nil, fmt.Errorf("INSERT statement missing relation")
	}

	// Build qualified table name with schema if present
	tableName := stmt.Relation.Relname
	if stmt.Relation.Schemaname != "" {
		tableName = stmt.Relation.Schemaname + "." + tableName
	}
	columns := make([]string, 0)

	// Extract column names if specified
	for _, target := range stmt.Cols {
		if resTarget := target.GetResTarget(); resTarget != nil {
			columns = append(columns, resTarget.Name)
		}
	}

	// Note: Extracting actual values from INSERT statements is complex
	// and not commonly needed for pg_dump parsing (which uses COPY for data)
	// This would be implemented if needed for specific use cases

	return &ParsedStatement{
		Type:      "INSERT",
		TableName: tableName,
		Columns:   columns,
		Metadata: map[string]interface{}{
			"has_values": stmt.SelectStmt != nil,
		},
	}, nil
}

// parseSelect parses a SELECT statement
func (p *SQLParser) parseSelect(_ *pg_query.SelectStmt) (*ParsedStatement, error) {
	// Check if this is a SELECT that calls a function (like setval)
	// These are important for sequence management in dumps

	// For now, we'll identify all SELECTs as SELECT type
	// The dump parser can check the SQL content to determine if it's a setval
	return &ParsedStatement{
		Type: "SELECT",
	}, nil
}

// NormalizeSQL normalizes a SQL statement for consistent processing
func (p *SQLParser) NormalizeSQL(sql string) string {
	// Remove extra whitespace and normalize case for keywords
	normalized := strings.TrimSpace(sql)

	// Convert common keywords to uppercase for consistency
	keywords := []string{"create", "table", "index", "alter", "drop", "insert", "select", "update", "delete"}
	for _, keyword := range keywords {
		re := regexp.MustCompile(`(?i)\b` + keyword + `\b`)
		normalized = re.ReplaceAllStringFunc(normalized, func(match string) string {
			return strings.ToUpper(match)
		})
	}

	return normalized
}

// ValidateSQL checks if a SQL statement is valid
func (p *SQLParser) ValidateSQL(sql string) error {
	_, err := pg_query.Parse(sql)
	return err
}

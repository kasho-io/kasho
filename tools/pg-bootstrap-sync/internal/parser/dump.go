package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// DumpParser implements the Parser interface for pg_dump files
type DumpParser struct {
	// Configuration options
	MaxRowsPerTable int // Limit rows per table for testing (0 = no limit)
}

// NewDumpParser creates a new dump parser
func NewDumpParser() *DumpParser {
	return &DumpParser{
		MaxRowsPerTable: 0, // No limit by default
	}
}

// Parse parses a pg_dump file from disk
func (p *DumpParser) Parse(filename string) (*ParseResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open dump file %s: %w", filename, err)
	}
	defer file.Close()

	result, err := p.ParseStream(file)
	if err != nil {
		return nil, err
	}

	result.Metadata.SourceFile = filename
	return result, nil
}

// ParseStream parses a pg_dump file from a reader
func (p *DumpParser) ParseStream(reader interface{}) (*ParseResult, error) {
	r, ok := reader.(io.Reader)
	if !ok {
		return nil, fmt.Errorf("reader must implement io.Reader")
	}

	scanner := bufio.NewScanner(r)
	
	result := &ParseResult{
		Statements: make([]Statement, 0),
		Metadata: ParseMetadata{
			ParsedAt:    time.Now(),
			TablesFound: make([]string, 0),
		},
	}

	var currentStatement strings.Builder
	var inCopyData bool
	var copyTable string
	var copyColumns []string
	var copyRows [][]string
	tableRowCounts := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "--") || strings.TrimSpace(line) == "" {
			continue
		}

		// Handle COPY statements
		if strings.HasPrefix(line, "COPY ") {
			// Parse COPY statement: COPY table (col1, col2, ...) FROM stdin;
			copyInfo := p.parseCopyStatement(line)
			if copyInfo != nil {
				inCopyData = true
				copyTable = copyInfo.table
				copyColumns = copyInfo.columns
				copyRows = make([][]string, 0)
				
				// Add table to found tables if not already present
				if !contains(result.Metadata.TablesFound, copyTable) {
					result.Metadata.TablesFound = append(result.Metadata.TablesFound, copyTable)
				}
			}
			continue
		}

		// Handle end of COPY data
		if line == "\\." {
			if inCopyData && len(copyRows) > 0 {
				// Create DML statement for the collected COPY data
				stmt := DMLStatement{
					Table:        copyTable,
					ColumnNames:  copyColumns,
					ColumnValues: copyRows,
				}
				result.Statements = append(result.Statements, stmt)
				result.Metadata.DMLCount++
			}
			inCopyData = false
			copyTable = ""
			copyColumns = nil
			copyRows = nil
			continue
		}

		// Handle COPY data rows
		if inCopyData {
			// Check row limit per table
			if p.MaxRowsPerTable > 0 {
				tableRowCounts[copyTable]++
				if tableRowCounts[copyTable] > p.MaxRowsPerTable {
					continue // Skip this row
				}
			}

			// Parse tab-separated values
			values := strings.Split(line, "\t")
			// Handle PostgreSQL COPY format escaping
			for i, val := range values {
				values[i] = p.unescapeCopyValue(val)
			}
			copyRows = append(copyRows, values)
			continue
		}

		// Handle regular SQL statements
		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		// Check if statement is complete (ends with semicolon)
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			sql := strings.TrimSpace(currentStatement.String())
			
			// Use proper SQL parser for all statements
			if err := p.parseWithSQLParser(sql, result); err != nil {
				// If SQL parser fails, log and skip the statement
				// This allows us to handle unsupported statement types gracefully
				continue
			}

			currentStatement.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading dump file: %w", err)
	}

	result.Metadata.StatementCount = len(result.Statements)
	return result, nil
}

// copyInfo holds parsed COPY statement information
type copyInfo struct {
	table   string
	columns []string
}

// parseCopyStatement parses a COPY statement and extracts table and column information
func (p *DumpParser) parseCopyStatement(line string) *copyInfo {
	// Match: COPY table_name (col1, col2, ...) FROM stdin;
	re := regexp.MustCompile(`COPY\s+(\w+)\s*\(([^)]+)\)\s+FROM\s+stdin;`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) != 3 {
		return nil
	}

	table := matches[1]
	columnsStr := matches[2]
	
	// Parse column names
	columns := make([]string, 0)
	for _, col := range strings.Split(columnsStr, ",") {
		columns = append(columns, strings.TrimSpace(col))
	}

	return &copyInfo{
		table:   table,
		columns: columns,
	}
}

// unescapeCopyValue unescapes PostgreSQL COPY format values
func (p *DumpParser) unescapeCopyValue(value string) string {
	if value == "\\N" {
		return "" // NULL value becomes empty string
	}
	
	// Unescape common PostgreSQL COPY escape sequences
	value = strings.ReplaceAll(value, "\\t", "\t")
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\r", "\r")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	
	return value
}

// parseWithSQLParser uses the proper SQL parser to parse statements
func (p *DumpParser) parseWithSQLParser(sql string, result *ParseResult) error {
	sqlParser := NewSQLParser()
	
	parsed, err := sqlParser.ParseSQL(sql)
	if err != nil {
		return err
	}
	
	for _, stmt := range parsed.Statements {
		switch stmt.Type {
		case "CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE":
			// Handle DDL statements
			ddlStmt := DDLStatement{
				SQL:      sql,
				Table:    stmt.TableName,
				Database: "unknown", // Could be extracted from dump header
				Time:     time.Now(),
			}
			result.Statements = append(result.Statements, ddlStmt)
			result.Metadata.DDLCount++
			
			// Add table to found tables if not already present
			if stmt.TableName != "" && !contains(result.Metadata.TablesFound, stmt.TableName) {
				result.Metadata.TablesFound = append(result.Metadata.TablesFound, stmt.TableName)
			}
			
		case "INSERT":
			// Handle INSERT statements - but we need to extract values
			// For now, fall back to manual parsing for INSERT values since
			// pg_query_go doesn't easily extract the actual data values
			insertInfo := p.parseInsertValues(sql)
			if insertInfo != nil {
				dmlStmt := DMLStatement{
					Table:        stmt.TableName,
					ColumnNames:  stmt.Columns, // Use columns from SQL parser
					ColumnValues: [][]string{insertInfo.values},
				}
				result.Statements = append(result.Statements, dmlStmt)
				result.Metadata.DMLCount++
				
				// Add table to found tables if not already present
				if !contains(result.Metadata.TablesFound, stmt.TableName) {
					result.Metadata.TablesFound = append(result.Metadata.TablesFound, stmt.TableName)
				}
			}
			
		default:
			// Skip unknown statement types
			continue
		}
	}
	
	return nil
}

// parseInsertValues extracts values from INSERT statements using simplified parsing
// This is a focused method that only handles the VALUES part, not the full SQL
func (p *DumpParser) parseInsertValues(sql string) *insertValues {
	upper := strings.ToUpper(sql)
	
	// Find where VALUES starts
	valuesIndex := strings.Index(upper, "VALUES")
	if valuesIndex == -1 {
		return nil
	}
	
	// Extract values section after VALUES keyword
	valuesSection := strings.TrimSpace(sql[valuesIndex+6:]) // 6 = len("VALUES")
	if !strings.HasPrefix(valuesSection, "(") {
		return nil
	}
	
	// Find matching closing parenthesis
	depth := 0
	var valuesEnd int
	for i, char := range valuesSection {
		if char == '(' {
			depth++
		} else if char == ')' {
			depth--
			if depth == 0 {
				valuesEnd = i
				break
			}
		}
	}
	
	if valuesEnd == 0 {
		return nil
	}
	
	// Extract values string (without outer parentheses)
	valuesStr := valuesSection[1:valuesEnd]
	
	// Parse values - split by comma, but respect quotes and parentheses
	values := p.parseValuesList(valuesStr)
	
	return &insertValues{
		values: values,
	}
}

// insertValues holds parsed INSERT values
type insertValues struct {
	values []string
}

// parseValuesList splits a comma-separated values list respecting quotes and parentheses
func (p *DumpParser) parseValuesList(valuesStr string) []string {
	values := make([]string, 0)
	current := strings.Builder{}
	inQuotes := false
	quoteChar := byte(0)
	depth := 0
	
	for i := 0; i < len(valuesStr); i++ {
		char := valuesStr[i]
		
		switch {
		case !inQuotes && (char == '\'' || char == '"'):
			// Start of quoted string
			inQuotes = true
			quoteChar = char
			current.WriteByte(char)
			
		case inQuotes && char == quoteChar:
			// End of quoted string (check for escaped quotes)
			if i+1 < len(valuesStr) && valuesStr[i+1] == quoteChar {
				// Escaped quote, include both
				current.WriteByte(char)
				current.WriteByte(char)
				i++ // Skip the next quote
			} else {
				// End of quotes
				inQuotes = false
				quoteChar = 0
				current.WriteByte(char)
			}
			
		case !inQuotes && char == '(':
			depth++
			current.WriteByte(char)
			
		case !inQuotes && char == ')':
			depth--
			current.WriteByte(char)
			
		case !inQuotes && depth == 0 && char == ',':
			// End of current value
			value := strings.TrimSpace(current.String())
			values = append(values, p.cleanValue(value))
			current.Reset()
			
		default:
			current.WriteByte(char)
		}
	}
	
	// Add the last value
	if current.Len() > 0 {
		value := strings.TrimSpace(current.String())
		values = append(values, p.cleanValue(value))
	}
	
	return values
}

// cleanValue removes quotes and handles NULL values
func (p *DumpParser) cleanValue(value string) string {
	value = strings.TrimSpace(value)
	
	// Handle NULL
	if strings.ToUpper(value) == "NULL" {
		return ""
	}
	
	// Remove quotes if present
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) ||
		   (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
			value = value[1 : len(value)-1]
			// Unescape basic SQL escapes
			value = strings.ReplaceAll(value, "''", "'")
			value = strings.ReplaceAll(value, "\"\"", "\"")
		}
	}
	
	return value
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
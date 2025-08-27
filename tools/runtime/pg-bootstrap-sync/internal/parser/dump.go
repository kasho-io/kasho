package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	var skipCopyData bool // Flag to skip COPY data for kasho_* tables
	var copyTable string
	var copyColumns []string
	var copyRows [][]string
	var inDollarQuote bool
	var dollarQuoteTag string
	tableRowCounts := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "--") || strings.TrimSpace(line) == "" {
			continue
		}

		// Skip psql meta-commands (like \restrict in PostgreSQL 17.6+)
		// These are client-side commands, not SQL statements
		// But don't skip the COPY data terminator \.
		if strings.HasPrefix(line, "\\") && line != "\\." {
			continue
		}

		// Handle COPY statements
		if strings.HasPrefix(line, "COPY ") {
			// Parse COPY statement: COPY table (col1, col2, ...) FROM stdin;
			copyInfo := p.parseCopyStatement(line)
			if copyInfo != nil {
				// Check if this is a kasho_* table
				tableParts := strings.Split(copyInfo.table, ".")
				tableName := tableParts[len(tableParts)-1]
				if strings.HasPrefix(tableName, "kasho_") {
					log.Printf("Skipping COPY data for Kasho internal table: %s", copyInfo.table)
					skipCopyData = true
					continue
				}
				
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
			if skipCopyData {
				// We were skipping kasho_* table data
				skipCopyData = false
				continue
			}
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
		if inCopyData || skipCopyData {
			if skipCopyData {
				// Skip lines until we see \.
				continue
			}
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

		// Check for dollar-quoted strings
		if !inDollarQuote {
			// Look for start of dollar quote
			if tag, startIdx := p.findDollarQuoteStart(line); tag != "" {
				// Check if the quote ends on the same line
				if endIdx := p.findDollarQuoteEnd(line[startIdx+len(tag):], tag); endIdx >= 0 {
					// Quote starts and ends on the same line, we're not inside a quote
					// Continue checking the rest of the line for more quotes
				} else {
					// Quote starts but doesn't end on this line
					inDollarQuote = true
					dollarQuoteTag = tag
				}
			}
		} else {
			// We're inside a dollar quote, look for the end
			if idx := p.findDollarQuoteEnd(line, dollarQuoteTag); idx >= 0 {
				inDollarQuote = false
				dollarQuoteTag = ""
			}
		}

		// Check if statement is complete (ends with semicolon and not inside dollar quote)
		if !inDollarQuote && strings.HasSuffix(strings.TrimSpace(line), ";") {
			sql := strings.TrimSpace(currentStatement.String())
			
			// Use proper SQL parser for all statements
			if err := p.parseWithSQLParser(sql, result); err != nil {
				// Fatal error for unsupported statements as requested
				return nil, fmt.Errorf("failed to parse statement: %w", err)
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
	// Use pg_query_go for robust parsing
	sqlParser := NewSQLParser()
	parsed, err := sqlParser.ParseSQL(line)
	if err != nil || len(parsed.Statements) == 0 {
		// Fall back to regex for simple cases if pg_query fails
		re := regexp.MustCompile(`COPY\s+([\w.]+)\s*\(([^)]+)\)\s+FROM\s+stdin;`)
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

	// Look for COPY statement in parsed results
	for _, stmt := range parsed.Statements {
		// Check if metadata indicates this is a COPY statement
		if stmt.Type == "COPY" || stmt.Type == "UNKNOWN" {
			// Extract table and columns from the parsed statement
			// For COPY statements, the table name should be in stmt.TableName
			if stmt.TableName != "" {
				// Build column list from stmt.Columns
				return &copyInfo{
					table:   stmt.TableName,
					columns: stmt.Columns,
				}
			}
		}
	}
	
	return nil
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
		// DML statements
		case "INSERT", "UPDATE", "DELETE":
			if stmt.Type == "INSERT" {
				// Skip INSERTs into kasho_* tables
				if stmt.TableName != "" {
					tableParts := strings.Split(stmt.TableName, ".")
					tableName := tableParts[len(tableParts)-1]
					if strings.HasPrefix(tableName, "kasho_") {
						log.Printf("Skipping INSERT into Kasho internal table: %s", stmt.TableName)
						continue
					}
				}
				
				// Handle INSERT statements - extract values
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
			} else {
				// UPDATE and DELETE are not expected in pg_dump but handle them as DML
				return fmt.Errorf("unexpected DML statement type %s in pg_dump", stmt.Type)
			}
			
		// DDL statements - all valid DDL that can appear in pg_dump
		case "CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE", "CREATE_SEQUENCE", "ALTER_SEQUENCE",
		     "CREATE_FUNCTION", "CREATE_TRIGGER", "CREATE_EVENT_TRIGGER", "DROP", "TRUNCATE", "COMMENT", "GRANT":
			// Check if this is a kasho_* internal object that should be skipped
			if p.isKashoInternalObject(stmt, sql) {
				log.Printf("Skipping Kasho internal object: %s", p.getObjectDescription(stmt, sql))
				continue
			}
			
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
			
		// Special handling for SELECT statements
		case "SELECT":
			// Check if this is a setval() call for sequences
			if strings.Contains(strings.ToUpper(sql), "PG_CATALOG.SETVAL") {
				// Check if this is for a kasho_* sequence
				if strings.Contains(sql, "kasho_") {
					log.Printf("Skipping setval for Kasho internal sequence")
					continue
				}
				// This is a sequence value setting - treat as DDL
				ddlStmt := DDLStatement{
					SQL:      sql,
					Table:    "", // No specific table for setval
					Database: "unknown",
					Time:     time.Now(),
				}
				result.Statements = append(result.Statements, ddlStmt)
				result.Metadata.DDLCount++
			} else if strings.Contains(strings.ToUpper(sql), "PG_CATALOG.SET_CONFIG") {
				// This is a session configuration setting - skip it like SET statements
				continue
			} else {
				// Other SELECT statements should not appear in pg_dump
				return fmt.Errorf("unexpected SELECT statement in pg_dump (not a setval): %s", sql)
			}
			
		// SET and TRANSACTION statements - skip these as they're session control
		case "SET", "TRANSACTION":
			// These are session control statements, skip them
			continue
			
		// Publication/Subscription statements - skip these as they're for replication setup only
		case "CREATE_PUBLICATION", "ALTER_PUBLICATION", "DROP_PUBLICATION",
		     "CREATE_SUBSCRIPTION", "ALTER_SUBSCRIPTION", "DROP_SUBSCRIPTION":
			// Skip publication/subscription statements
			continue
			
		// Unknown statement type
		case "UNKNOWN":
			
			// Get the actual node type from metadata for better error message
			nodeType := "unknown"
			if stmt.Metadata != nil {
				if nt, ok := stmt.Metadata["node_type"].(string); ok {
					nodeType = nt
				}
			}
			return fmt.Errorf("unsupported statement type in pg_dump: %s (node type: %s), SQL: %s", stmt.Type, nodeType, sql)
			
		default:
			// Any other statement type is an error
			return fmt.Errorf("unsupported statement type in pg_dump: %s, SQL: %s", stmt.Type, sql)
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

// findDollarQuoteStart finds the start of a dollar-quoted string in a line
// Returns the tag (e.g., "$$" or "$func$") and its starting index if found
func (p *DumpParser) findDollarQuoteStart(line string) (string, int) {
	// Regular expression to match dollar quote tags: $[tag]$
	// Tag can be empty ($$) or contain alphanumeric/underscore ($tag$)
	re := regexp.MustCompile(`\$([A-Za-z_]\w*)?\$`)
	loc := re.FindStringIndex(line)
	if loc != nil {
		tag := line[loc[0]:loc[1]]
		return tag, loc[0]
	}
	return "", -1
}

// findDollarQuoteEnd finds the end of a specific dollar-quoted string in a line
// Returns the index of the closing tag if found, -1 otherwise
func (p *DumpParser) findDollarQuoteEnd(line, tag string) int {
	// Look for the closing tag
	return strings.Index(line, tag)
}

// isKashoInternalObject checks if a statement creates/modifies a kasho_* internal object
func (p *DumpParser) isKashoInternalObject(stmt ParsedStatement, sql string) bool {
	// Check table name for kasho_ prefix (handle schema.table format)
	if stmt.TableName != "" {
		// Extract table name without schema
		tableParts := strings.Split(stmt.TableName, ".")
		tableName := tableParts[len(tableParts)-1]
		if strings.HasPrefix(tableName, "kasho_") {
			return true
		}
	}
	
	// For functions, triggers, and event triggers, we need to check the SQL
	upperSQL := strings.ToUpper(sql)
	
	switch stmt.Type {
	case "CREATE_FUNCTION":
		// Check for kasho_* function names
		if strings.Contains(upperSQL, "FUNCTION KASHO_") ||
		   strings.Contains(upperSQL, "FUNCTION PUBLIC.KASHO_") {
			return true
		}
	case "CREATE_TRIGGER":
		// Check for kasho_* trigger names
		if strings.Contains(upperSQL, "TRIGGER KASHO_") {
			return true
		}
	case "CREATE_EVENT_TRIGGER":
		// Check for kasho_* event trigger names
		if strings.Contains(upperSQL, "EVENT TRIGGER KASHO_") {
			return true
		}
	case "ALTER_TABLE", "DROP", "COMMENT", "GRANT":
		// Check if operating on kasho_* objects
		if strings.Contains(upperSQL, " KASHO_") ||
		   strings.Contains(upperSQL, ".KASHO_") {
			return true
		}
	}
	
	return false
}

// getObjectDescription returns a human-readable description of the object being skipped
func (p *DumpParser) getObjectDescription(stmt ParsedStatement, sql string) string {
	switch stmt.Type {
	case "CREATE_TABLE":
		return fmt.Sprintf("table %s", stmt.TableName)
	case "CREATE_FUNCTION":
		// Extract function name from SQL
		re := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+(?:[\w.]+\.)?(kasho_\w+)`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return fmt.Sprintf("function %s", matches[1])
		}
		return "kasho_* function"
	case "CREATE_TRIGGER":
		// Extract trigger name from SQL
		re := regexp.MustCompile(`(?i)CREATE\s+TRIGGER\s+(kasho_\w+)`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return fmt.Sprintf("trigger %s", matches[1])
		}
		return "kasho_* trigger"
	case "CREATE_EVENT_TRIGGER":
		// Extract event trigger name from SQL
		re := regexp.MustCompile(`(?i)CREATE\s+EVENT\s+TRIGGER\s+(kasho_\w+)`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return fmt.Sprintf("event trigger %s", matches[1])
		}
		return "kasho_* event trigger"
	default:
		return fmt.Sprintf("%s on kasho_* object", strings.ToLower(stmt.Type))
	}
}
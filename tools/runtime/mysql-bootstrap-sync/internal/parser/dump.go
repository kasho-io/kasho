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

// DumpParser implements the Parser interface for mysqldump files
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

// Parse parses a mysqldump file from disk
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

// ParseStream parses a mysqldump file from a reader
func (p *DumpParser) ParseStream(reader interface{}) (*ParseResult, error) {
	r, ok := reader.(io.Reader)
	if !ok {
		return nil, fmt.Errorf("reader must implement io.Reader")
	}

	scanner := bufio.NewScanner(r)
	// Increase buffer size for large INSERT statements (extended inserts can be very long)
	const maxScanTokenSize = 64 * 1024 * 1024 // 64MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	result := &ParseResult{
		Statements: make([]Statement, 0),
		Metadata: ParseMetadata{
			ParsedAt:    time.Now(),
			TablesFound: make([]string, 0),
		},
	}

	var currentStatement strings.Builder
	tableRowCounts := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip MySQL comments and empty lines
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" ||
			strings.HasPrefix(trimmedLine, "--") ||
			strings.HasPrefix(trimmedLine, "/*") ||
			strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Skip SET statements and session control
		upperLine := strings.ToUpper(trimmedLine)
		if strings.HasPrefix(upperLine, "SET ") ||
			strings.HasPrefix(upperLine, "LOCK TABLES") ||
			strings.HasPrefix(upperLine, "UNLOCK TABLES") ||
			strings.HasPrefix(upperLine, "START TRANSACTION") ||
			strings.HasPrefix(upperLine, "COMMIT") ||
			strings.HasPrefix(upperLine, "USE ") {
			continue
		}

		// Build up multi-line statements
		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		// Check if statement is complete (ends with semicolon)
		if strings.HasSuffix(trimmedLine, ";") {
			sql := strings.TrimSpace(currentStatement.String())
			currentStatement.Reset()

			// Parse the complete statement
			if err := p.parseStatement(sql, result, tableRowCounts); err != nil {
				log.Printf("Warning: failed to parse statement: %v", err)
				// Continue parsing - don't fail on individual statement errors
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading dump file: %w", err)
	}

	result.Metadata.StatementCount = len(result.Statements)
	return result, nil
}

// parseStatement parses a single SQL statement and adds it to the result
func (p *DumpParser) parseStatement(sql string, result *ParseResult, tableRowCounts map[string]int) error {
	upperSQL := strings.ToUpper(sql)

	// Handle INSERT statements (DML)
	if strings.HasPrefix(upperSQL, "INSERT ") {
		return p.parseInsertStatement(sql, result, tableRowCounts)
	}

	// Handle DDL statements
	if strings.HasPrefix(upperSQL, "CREATE ") ||
		strings.HasPrefix(upperSQL, "ALTER ") ||
		strings.HasPrefix(upperSQL, "DROP ") ||
		strings.HasPrefix(upperSQL, "TRUNCATE ") {
		return p.parseDDLStatement(sql, result)
	}

	// Skip other statement types (REPLACE, UPDATE, DELETE not expected in mysqldump)
	return nil
}

// parseInsertStatement parses an INSERT statement from mysqldump
func (p *DumpParser) parseInsertStatement(sql string, result *ParseResult, tableRowCounts map[string]int) error {
	// Extract table name: INSERT INTO `table` or INSERT INTO table
	tableMatch := regexp.MustCompile(`(?i)INSERT\s+INTO\s+` + "`?" + `([\w.]+)` + "`?" + `\s*`).FindStringSubmatch(sql)
	if len(tableMatch) < 2 {
		return fmt.Errorf("could not extract table name from INSERT")
	}
	tableName := tableMatch[1]

	// Skip kasho_* internal tables
	if strings.HasPrefix(strings.ToLower(tableName), "kasho_") {
		log.Printf("Skipping INSERT into Kasho internal table: %s", tableName)
		return nil
	}

	// Extract column names if present: INSERT INTO table (col1, col2, ...)
	var columnNames []string
	columnsMatch := regexp.MustCompile(`(?i)INSERT\s+INTO\s+` + "`?" + `[\w.]+` + "`?" + `\s*\(([^)]+)\)`).FindStringSubmatch(sql)
	if len(columnsMatch) >= 2 {
		columnsStr := columnsMatch[1]
		for _, col := range strings.Split(columnsStr, ",") {
			colName := strings.TrimSpace(col)
			colName = strings.Trim(colName, "`\"")
			columnNames = append(columnNames, colName)
		}
	}

	// Extract VALUES section
	valuesIndex := strings.Index(strings.ToUpper(sql), "VALUES")
	if valuesIndex == -1 {
		return fmt.Errorf("no VALUES clause found in INSERT")
	}

	valuesSection := strings.TrimSpace(sql[valuesIndex+6:]) // 6 = len("VALUES")
	valuesSection = strings.TrimSuffix(valuesSection, ";")

	// Parse multiple value sets (extended inserts): (v1,v2),(v3,v4),...
	valueSets := p.parseValueSets(valuesSection)

	// Apply max rows per table limit
	var rows [][]string
	for _, valueSet := range valueSets {
		if p.MaxRowsPerTable > 0 {
			tableRowCounts[tableName]++
			if tableRowCounts[tableName] > p.MaxRowsPerTable {
				continue
			}
		}
		rows = append(rows, valueSet)
	}

	if len(rows) == 0 {
		return nil
	}

	// Add table to found tables if not already present
	if !contains(result.Metadata.TablesFound, tableName) {
		result.Metadata.TablesFound = append(result.Metadata.TablesFound, tableName)
	}

	// Create DML statement
	stmt := DMLStatement{
		Table:        tableName,
		ColumnNames:  columnNames,
		ColumnValues: rows,
	}
	result.Statements = append(result.Statements, stmt)
	result.Metadata.DMLCount++

	return nil
}

// parseValueSets parses the VALUES section containing multiple value sets
func (p *DumpParser) parseValueSets(valuesSection string) [][]string {
	var valueSets [][]string
	var currentSet []string
	var currentValue strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	depth := 0

	for i := 0; i < len(valuesSection); i++ {
		char := valuesSection[i]

		switch {
		case !inQuotes && char == '(':
			depth++
			if depth == 1 {
				// Start of a new value set
				currentSet = nil
				currentValue.Reset()
				continue
			}
			currentValue.WriteByte(char)

		case !inQuotes && char == ')':
			depth--
			if depth == 0 {
				// End of current value set
				if currentValue.Len() > 0 || len(currentSet) > 0 {
					currentSet = append(currentSet, p.cleanMySQLValue(currentValue.String()))
					valueSets = append(valueSets, currentSet)
				}
				currentSet = nil
				currentValue.Reset()
				continue
			}
			currentValue.WriteByte(char)

		case !inQuotes && (char == '\'' || char == '"'):
			inQuotes = true
			quoteChar = char
			currentValue.WriteByte(char)

		case inQuotes && char == quoteChar:
			// Check for escaped quote
			if i+1 < len(valuesSection) && valuesSection[i+1] == quoteChar {
				currentValue.WriteByte(char)
				currentValue.WriteByte(char)
				i++
			} else if char == '\'' && i > 0 && valuesSection[i-1] == '\\' {
				// Escaped quote with backslash (already written)
				currentValue.WriteByte(char)
			} else {
				inQuotes = false
				quoteChar = 0
				currentValue.WriteByte(char)
			}

		case !inQuotes && depth == 1 && char == ',':
			// Value separator within a value set
			currentSet = append(currentSet, p.cleanMySQLValue(currentValue.String()))
			currentValue.Reset()

		default:
			if depth > 0 {
				currentValue.WriteByte(char)
			}
		}
	}

	return valueSets
}

// cleanMySQLValue cleans a MySQL value from the dump
func (p *DumpParser) cleanMySQLValue(value string) string {
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
			// Unescape MySQL escapes
			value = strings.ReplaceAll(value, "''", "'")
			value = strings.ReplaceAll(value, "\\\\", "\\")
			value = strings.ReplaceAll(value, "\\'", "'")
			value = strings.ReplaceAll(value, "\\\"", "\"")
			value = strings.ReplaceAll(value, "\\n", "\n")
			value = strings.ReplaceAll(value, "\\r", "\r")
			value = strings.ReplaceAll(value, "\\t", "\t")
			value = strings.ReplaceAll(value, "\\0", "\x00")
		}
	}

	return value
}

// parseDDLStatement parses a DDL statement
func (p *DumpParser) parseDDLStatement(sql string, result *ParseResult) error {
	// Extract table name from DDL
	tableName := p.extractTableName(sql)

	// Skip kasho_* internal tables
	if tableName != "" && strings.HasPrefix(strings.ToLower(tableName), "kasho_") {
		log.Printf("Skipping DDL for Kasho internal table: %s", tableName)
		return nil
	}

	stmt := DDLStatement{
		SQL:      sql,
		Table:    tableName,
		Database: "unknown",
		Time:     time.Now(),
	}
	result.Statements = append(result.Statements, stmt)
	result.Metadata.DDLCount++

	// Add table to found tables if not already present
	if tableName != "" && !contains(result.Metadata.TablesFound, tableName) {
		result.Metadata.TablesFound = append(result.Metadata.TablesFound, tableName)
	}

	return nil
}

// extractTableName extracts the table name from a DDL statement
func (p *DumpParser) extractTableName(sql string) string {
	// Try various patterns for extracting table name
	patterns := []string{
		`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?` + "`?" + `([\w.]+)` + "`?",
		`(?i)ALTER\s+TABLE\s+` + "`?" + `([\w.]+)` + "`?",
		`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?` + "`?" + `([\w.]+)` + "`?",
		`(?i)TRUNCATE\s+(?:TABLE\s+)?` + "`?" + `([\w.]+)` + "`?",
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(sql)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
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

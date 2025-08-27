package parser

import (
	"time"
)

// Statement represents a parsed SQL statement from a pg_dump file
type Statement interface {
	Type() string
}

// DDLStatement represents a Data Definition Language statement
type DDLStatement struct {
	SQL      string    // The raw SQL statement
	Table    string    // Table name if applicable
	Database string    // Database name
	Time     time.Time // When the statement was parsed/executed
}

func (d DDLStatement) Type() string {
	return "ddl"
}

// DMLStatement represents a Data Manipulation Language statement (typically COPY data)
type DMLStatement struct {
	Table        string     // Table name
	ColumnNames  []string   // Column names from COPY statement
	ColumnValues [][]string // Rows of data (each row is []string of values)
}

func (d DMLStatement) Type() string {
	return "dml"
}

// ParseResult represents the result of parsing a pg_dump file
type ParseResult struct {
	Statements []Statement
	Metadata   ParseMetadata
}

// ParseMetadata contains metadata about the parsing process
type ParseMetadata struct {
	SourceFile     string    // Path to the source dump file
	ParsedAt       time.Time // When parsing completed
	StatementCount int       // Total number of statements parsed
	DDLCount       int       // Number of DDL statements
	DMLCount       int       // Number of DML statements
	TablesFound    []string  // List of table names encountered
}

// Parser defines the interface for parsing pg_dump files
type Parser interface {
	Parse(filename string) (*ParseResult, error)
	ParseStream(reader interface{}) (*ParseResult, error)
}

// DumpFormat represents the format of the pg_dump file
type DumpFormat int

const (
	FormatPlainText DumpFormat = iota
	FormatCustom
	FormatTar
	FormatDirectory
)

// DumpInfo contains information about the dump file
type DumpInfo struct {
	Format            DumpFormat
	PostgreSQLVersion string
	DumpedAt          time.Time
	DatabaseName      string
}

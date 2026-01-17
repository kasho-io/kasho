package parser

import "time"

// Statement is an interface for parsed SQL statements
type Statement interface {
	GetType() string
}

// DDLStatement represents a DDL statement
type DDLStatement struct {
	SQL      string
	Table    string
	Database string
	Time     time.Time
}

func (d DDLStatement) GetType() string {
	return "DDL"
}

// DMLStatement represents a DML statement (INSERT in mysqldump)
type DMLStatement struct {
	Table        string
	ColumnNames  []string
	ColumnValues [][]string // Multiple rows from extended inserts
}

func (d DMLStatement) GetType() string {
	return "DML"
}

// ParseResult holds the result of parsing a dump file
type ParseResult struct {
	Statements []Statement
	Metadata   ParseMetadata
}

// ParseMetadata contains metadata about the parsed dump
type ParseMetadata struct {
	SourceFile     string
	ParsedAt       time.Time
	StatementCount int
	DDLCount       int
	DMLCount       int
	TablesFound    []string
}

// Parser interface for dump file parsers
type Parser interface {
	Parse(filename string) (*ParseResult, error)
}

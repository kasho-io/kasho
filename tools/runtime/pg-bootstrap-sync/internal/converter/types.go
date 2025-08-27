package converter

import (
	"kasho/proto"
)

// ConversionResult represents the result of converting parsed statements to changes
type ConversionResult struct {
	Changes     []*proto.Change
	LastLSN     string
	ChangeCount int
	DDLCount    int
	DMLCount    int
	Statistics  ConversionStatistics
}

// ConversionStatistics contains statistics about the conversion process
type ConversionStatistics struct {
	TablesProcessed   int
	RowsProcessed     int64
	BytesProcessed    int64
	ConversionTime    int64 // Duration in milliseconds
	ErrorsEncountered int
}

// ConversionConfig contains configuration for the conversion process
type ConversionConfig struct {
	StartingSequence int64    // Starting sequence number for LSN generation
	SnapshotLSN      string   // The LSN where WAL replication will begin
	MaxRowsPerBatch  int      // Maximum rows to process per batch (0 = no limit)
	SkipTables       []string // Tables to skip during conversion
	OnlyTables       []string // Only process these tables (empty = process all)
}

// DefaultConversionConfig returns a default conversion configuration
func DefaultConversionConfig() ConversionConfig {
	return ConversionConfig{
		StartingSequence: 0,
		MaxRowsPerBatch:  10000,
		SkipTables:       []string{},
		OnlyTables:       []string{},
	}
}

package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"kasho/pkg/kvbuffer"
	"kasho/proto"
	"pg-bootstrap-sync/internal/converter"
	"pg-bootstrap-sync/internal/parser"
)

// Bootstrapper orchestrates the bootstrap process
type Bootstrapper struct {
	parser    parser.Parser
	converter *converter.ChangeConverter
	kvBuffer  *kvbuffer.KVBuffer
	config    Config
	stats     Statistics
}

// Config contains configuration for the bootstrap process
type Config struct {
	DumpFile          string
	KVBufferURL       string
	BatchSize         int
	MaxRowsPerTable   int
	ProgressInterval  int // Log progress every N changes
	ResumeFromLSN     string
	DryRun           bool
}

// Statistics tracks bootstrap progress
type Statistics struct {
	StartTime        time.Time
	EndTime          time.Time
	StatementsRead   int
	ChangesGenerated int
	ChangesStored    int
	DDLCount         int
	DMLCount         int
	TablesProcessed  []string
	LastLSN          string
	BytesProcessed   int64
	ErrorsEncountered int
}

// NewBootstrapper creates a new bootstrapper instance
func NewBootstrapper(config Config) (*Bootstrapper, error) {
	// Create parser
	dumpParser := parser.NewDumpParser()
	if config.MaxRowsPerTable > 0 {
		dumpParser.MaxRowsPerTable = config.MaxRowsPerTable
	}

	// Create converter
	conv := converter.NewChangeConverter()

	// Create KV buffer connection
	var kvBuffer *kvbuffer.KVBuffer
	if !config.DryRun {
		var err error
		kvBuffer, err = kvbuffer.NewKVBuffer(config.KVBufferURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to KV buffer: %w", err)
		}
	}

	return &Bootstrapper{
		parser:    dumpParser,
		converter: conv,
		kvBuffer:  kvBuffer,
		config:    config,
		stats:     Statistics{},
	}, nil
}

// Bootstrap executes the full bootstrap process
func (b *Bootstrapper) Bootstrap(ctx context.Context) error {
	b.stats.StartTime = time.Now()
	slog.Info("Starting bootstrap process",
		"dump_file", b.config.DumpFile)

	// Parse the dump file
	slog.Info("Parsing dump file")
	parseResult, err := b.parser.Parse(b.config.DumpFile)
	if err != nil {
		return fmt.Errorf("failed to parse dump file: %w", err)
	}

	b.stats.StatementsRead = len(parseResult.Statements)
	slog.Info("Parsed statements successfully",
		"total_statements", parseResult.Metadata.StatementCount,
		"ddl_count", parseResult.Metadata.DDLCount,
		"dml_count", parseResult.Metadata.DMLCount,
		"tables_found", len(parseResult.Metadata.TablesFound))

	// Convert statements to changes
	slog.Info("Converting statements to changes")
	changes, err := b.converter.ConvertStatements(parseResult.Statements)
	if err != nil {
		return fmt.Errorf("failed to convert statements: %w", err)
	}

	b.stats.ChangesGenerated = len(changes)
	slog.Info("Generated changes",
		"change_count", len(changes))

	// Store changes in KV buffer
	if !b.config.DryRun {
		slog.Info("Storing changes in KV buffer",
			"batch_size", b.config.BatchSize)
		err = b.storeChanges(ctx, changes)
		if err != nil {
			return fmt.Errorf("failed to store changes: %w", err)
		}
	} else {
		slog.Info("Dry run mode - skipping KV buffer storage")
		b.stats.ChangesStored = len(changes)
	}

	// Update final statistics
	b.stats.EndTime = time.Now()
	if len(changes) > 0 {
		b.stats.LastLSN = changes[len(changes)-1].Lsn
	}
	b.stats.TablesProcessed = parseResult.Metadata.TablesFound
	b.stats.DDLCount = parseResult.Metadata.DDLCount
	b.stats.DMLCount = parseResult.Metadata.DMLCount

	b.logFinalStatistics()
	return nil
}

// storeChanges stores changes in the KV buffer with progress tracking
func (b *Bootstrapper) storeChanges(ctx context.Context, changes []*proto.Change) error {
	batchSize := b.config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	progressInterval := b.config.ProgressInterval
	if progressInterval <= 0 {
		progressInterval = 1000 // Default progress interval
	}

	stored := 0
	for i, change := range changes {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Store the change
		err := b.kvBuffer.AddChange(ctx, &BootstrapChange{change})
		if err != nil {
			b.stats.ErrorsEncountered++
			slog.Error("Failed to store change",
				"change_index", i+1,
				"lsn", change.Lsn,
				"error", err)
			continue
		}

		stored++
		b.stats.ChangesStored = stored

		// Log progress
		if (i+1)%progressInterval == 0 {
			elapsed := time.Since(b.stats.StartTime)
			rate := float64(stored) / elapsed.Seconds()
			slog.Info("Storage progress",
				"stored", stored,
				"total", len(changes),
				"rate", fmt.Sprintf("%.1f changes/sec", rate),
				"percentage", fmt.Sprintf("%.1f%%", float64(stored)/float64(len(changes))*100))
		}

		// Optional: Add small delay between batches to avoid overwhelming Redis
		if (i+1)%batchSize == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	slog.Info("Storage completed",
		"stored", stored,
		"total", len(changes),
		"success_rate", fmt.Sprintf("%.1f%%", float64(stored)/float64(len(changes))*100))
	return nil
}

// logFinalStatistics logs final bootstrap statistics
func (b *Bootstrapper) logFinalStatistics() {
	duration := b.stats.EndTime.Sub(b.stats.StartTime)
	
	logLevel := slog.LevelInfo
	if b.stats.ErrorsEncountered > 0 {
		logLevel = slog.LevelWarn
	}
	
	logFields := []interface{}{
		"duration", duration,
		"statements_read", b.stats.StatementsRead,
		"changes_generated", b.stats.ChangesGenerated,
		"changes_stored", b.stats.ChangesStored,
		"ddl_count", b.stats.DDLCount,
		"dml_count", b.stats.DMLCount,
		"tables_processed_count", len(b.stats.TablesProcessed),
		"tables_list", b.stats.TablesProcessed,
		"last_lsn", b.stats.LastLSN,
		"errors_encountered", b.stats.ErrorsEncountered,
	}
	
	if b.stats.ChangesStored > 0 && duration.Seconds() > 0 {
		rate := float64(b.stats.ChangesStored) / duration.Seconds()
		logFields = append(logFields, "average_rate", fmt.Sprintf("%.1f changes/sec", rate))
	}
	
	if b.stats.ErrorsEncountered > 0 {
		slog.Log(nil, logLevel, "Bootstrap completed with errors", logFields...)
	} else {
		slog.Log(nil, logLevel, "Bootstrap completed successfully", logFields...)
	}
}

// GetStatistics returns the current bootstrap statistics
func (b *Bootstrapper) GetStatistics() Statistics {
	return b.stats
}

// Close closes the bootstrapper and cleans up resources
func (b *Bootstrapper) Close() error {
	if b.kvBuffer != nil {
		return b.kvBuffer.Close()
	}
	return nil
}

// BootstrapChange wraps a proto.Change to implement the kvbuffer.Change interface
type BootstrapChange struct {
	*proto.Change
}

// Type implements kvbuffer.Change interface
func (bc *BootstrapChange) Type() string {
	return bc.Change.Type
}

// GetLSN implements kvbuffer.Change interface  
func (bc *BootstrapChange) GetLSN() string {
	return bc.Change.Lsn
}
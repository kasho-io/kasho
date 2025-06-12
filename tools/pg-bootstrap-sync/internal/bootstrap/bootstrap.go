package bootstrap

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("Starting bootstrap process for dump file: %s", b.config.DumpFile)

	// Parse the dump file
	log.Println("Parsing dump file...")
	parseResult, err := b.parser.Parse(b.config.DumpFile)
	if err != nil {
		return fmt.Errorf("failed to parse dump file: %w", err)
	}

	b.stats.StatementsRead = len(parseResult.Statements)
	log.Printf("Parsed %d statements (%d DDL, %d DML)", 
		parseResult.Metadata.StatementCount, 
		parseResult.Metadata.DDLCount, 
		parseResult.Metadata.DMLCount)

	// Convert statements to changes
	log.Println("Converting statements to changes...")
	changes, err := b.converter.ConvertStatements(parseResult.Statements)
	if err != nil {
		return fmt.Errorf("failed to convert statements: %w", err)
	}

	b.stats.ChangesGenerated = len(changes)
	log.Printf("Generated %d changes", len(changes))

	// Store changes in KV buffer
	if !b.config.DryRun {
		log.Println("Storing changes in KV buffer...")
		err = b.storeChanges(ctx, changes)
		if err != nil {
			return fmt.Errorf("failed to store changes: %w", err)
		}
	} else {
		log.Println("Dry run mode: skipping KV buffer storage")
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
			log.Printf("Failed to store change %d (LSN: %s): %v", i+1, change.Lsn, err)
			continue
		}

		stored++
		b.stats.ChangesStored = stored

		// Log progress
		if (i+1)%progressInterval == 0 {
			elapsed := time.Since(b.stats.StartTime)
			rate := float64(stored) / elapsed.Seconds()
			log.Printf("Stored %d/%d changes (%.1f changes/sec)", stored, len(changes), rate)
		}

		// Optional: Add small delay between batches to avoid overwhelming Redis
		if (i+1)%batchSize == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	log.Printf("Successfully stored %d/%d changes", stored, len(changes))
	return nil
}

// logFinalStatistics logs final bootstrap statistics
func (b *Bootstrapper) logFinalStatistics() {
	duration := b.stats.EndTime.Sub(b.stats.StartTime)
	
	log.Println("Bootstrap completed successfully!")
	log.Printf("Duration: %v", duration)
	log.Printf("Statements read: %d", b.stats.StatementsRead)
	log.Printf("Changes generated: %d", b.stats.ChangesGenerated)
	log.Printf("Changes stored: %d", b.stats.ChangesStored)
	log.Printf("DDL changes: %d", b.stats.DDLCount)
	log.Printf("DML changes: %d", b.stats.DMLCount)
	log.Printf("Tables processed: %d (%v)", len(b.stats.TablesProcessed), b.stats.TablesProcessed)
	log.Printf("Last LSN: %s", b.stats.LastLSN)
	log.Printf("Errors encountered: %d", b.stats.ErrorsEncountered)
	
	if b.stats.ChangesStored > 0 && duration.Seconds() > 0 {
		rate := float64(b.stats.ChangesStored) / duration.Seconds()
		log.Printf("Average rate: %.1f changes/sec", rate)
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
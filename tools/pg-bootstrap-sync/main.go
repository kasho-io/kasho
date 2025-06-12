package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"pg-bootstrap-sync/internal/bootstrap"
)

var (
	dumpFile         string
	kvURL            string
	batchSize        int
	maxRowsPerTable  int
	progressInterval int
	dryRun           bool
	verbose          bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pg-bootstrap-sync",
		Short: "Bootstrap PostgreSQL replica databases from pg_dump files",
		Long: `pg-bootstrap-sync parses PostgreSQL dump files and converts them to Change objects
that can be consumed by the Kasho replication infrastructure. This enables bootstrapping
replica databases with historical data before starting real-time WAL replication.`,
		RunE: runBootstrap,
	}

	// Define command-line flags
	rootCmd.Flags().StringVarP(&dumpFile, "dump-file", "d", "", "Path to pg_dump file (required)")
	rootCmd.Flags().StringVarP(&kvURL, "kv-url", "k", "", "Redis connection URL (required)")
	rootCmd.Flags().IntVarP(&batchSize, "batch-size", "b", 1000, "Processing batch size")
	rootCmd.Flags().IntVarP(&maxRowsPerTable, "max-rows-per-table", "m", 0, "Maximum rows per table (0 = no limit)")
	rootCmd.Flags().IntVarP(&progressInterval, "progress-interval", "p", 1000, "Log progress every N changes")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Parse and convert but don't store in KV buffer")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Mark required flags
	rootCmd.MarkFlagRequired("dump-file")

	// Only require kv-url if not doing a dry run
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if !dryRun && kvURL == "" {
			return fmt.Errorf("--kv-url is required unless --dry-run is specified")
		}
		return nil
	}

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	// Set up logging
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	// Validate dump file exists
	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		return fmt.Errorf("dump file does not exist: %s", dumpFile)
	}

	// Create bootstrap configuration
	config := bootstrap.Config{
		DumpFile:          dumpFile,
		KVBufferURL:       kvURL,
		BatchSize:         batchSize,
		MaxRowsPerTable:   maxRowsPerTable,
		ProgressInterval:  progressInterval,
		DryRun:           dryRun,
	}

	// Create bootstrapper
	bootstrapper, err := bootstrap.NewBootstrapper(config)
	if err != nil {
		return fmt.Errorf("failed to create bootstrapper: %w", err)
	}
	defer bootstrapper.Close()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, cancelling bootstrap...")
		cancel()
	}()

	// Run the bootstrap process
	err = bootstrapper.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Log final statistics
	stats := bootstrapper.GetStatistics()
	if stats.ErrorsEncountered > 0 {
		log.Printf("Bootstrap completed with %d errors", stats.ErrorsEncountered)
		return fmt.Errorf("bootstrap completed with errors")
	}

	log.Println("Bootstrap completed successfully!")
	return nil
}
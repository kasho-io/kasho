package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"kasho/pkg/version"
	"mysql-bootstrap-sync/internal/bootstrap"
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
		Use:   "mysql-bootstrap-sync",
		Short: "Bootstrap MySQL replica databases from mysqldump files",
		Long: `mysql-bootstrap-sync parses MySQL dump files and converts them to Change objects
that can be consumed by the Kasho replication infrastructure. This enables bootstrapping
replica databases with historical data before starting real-time binlog replication.`,
		RunE: runBootstrap,
	}

	// Define command-line flags
	rootCmd.Flags().StringVarP(&dumpFile, "dump-file", "d", "", "Path to mysqldump file (required)")
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	// Set up structured logging
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: verbose,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Log startup configuration
	slog.Info("Starting mysql-bootstrap-sync",
		"version", version.Version,
		"commit", version.GitCommit,
		"built", version.BuildDate,
		"dump_file", dumpFile,
		"batch_size", batchSize,
		"max_rows_per_table", maxRowsPerTable,
		"progress_interval", progressInterval,
		"dry_run", dryRun,
		"verbose", verbose,
	)

	// Validate dump file exists
	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		slog.Error("Dump file does not exist", "path", dumpFile)
		return fmt.Errorf("dump file does not exist: %s", dumpFile)
	}

	// Create bootstrap configuration
	config := bootstrap.Config{
		DumpFile:         dumpFile,
		KVBufferURL:      kvURL,
		BatchSize:        batchSize,
		MaxRowsPerTable:  maxRowsPerTable,
		ProgressInterval: progressInterval,
		DryRun:           dryRun,
	}

	// Create bootstrapper
	bootstrapper, err := bootstrap.NewBootstrapper(config)
	if err != nil {
		return fmt.Errorf("failed to create bootstrapper: %w", err)
	}
	defer bootstrapper.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Warn("Received shutdown signal, cancelling bootstrap...")
		cancel()
	}()

	// Run the bootstrap process
	err = bootstrapper.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	// Log final statistics
	stats := bootstrapper.GetStatistics()

	duration := stats.EndTime.Sub(stats.StartTime)

	slog.Info("Bootstrap completed",
		"duration", duration,
		"statements_read", stats.StatementsRead,
		"changes_generated", stats.ChangesGenerated,
		"changes_stored", stats.ChangesStored,
		"ddl_count", stats.DDLCount,
		"dml_count", stats.DMLCount,
		"tables_processed", len(stats.TablesProcessed),
		"errors_encountered", stats.ErrorsEncountered,
		"average_rate", fmt.Sprintf("%.1f changes/sec", float64(stats.ChangesStored)/duration.Seconds()),
	)

	if stats.ErrorsEncountered > 0 {
		slog.Error("Bootstrap completed with errors", "error_count", stats.ErrorsEncountered)
		return fmt.Errorf("bootstrap completed with %d errors", stats.ErrorsEncountered)
	}

	return nil
}

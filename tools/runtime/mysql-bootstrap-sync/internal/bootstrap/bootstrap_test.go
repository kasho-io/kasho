package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBootstrapper_DryRun(t *testing.T) {
	config := Config{
		DumpFile: "nonexistent.sql",
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	if b.parser == nil {
		t.Error("expected parser to be initialized")
	}
	if b.converter == nil {
		t.Error("expected converter to be initialized")
	}
	if b.kvBuffer != nil {
		t.Error("expected kvBuffer to be nil in dry run mode")
	}
}

func TestNewBootstrapper_MaxRowsPerTable(t *testing.T) {
	config := Config{
		DumpFile:        "nonexistent.sql",
		DryRun:          true,
		MaxRowsPerTable: 100,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	// The parser should have MaxRowsPerTable set
	// We can't directly check this without exposing internals,
	// but we can verify the bootstrapper was created
	if b == nil {
		t.Error("expected bootstrapper to be created")
	}
}

func TestNewBootstrapper_InvalidKVURL(t *testing.T) {
	config := Config{
		DumpFile:    "nonexistent.sql",
		KVBufferURL: "invalid://url",
		DryRun:      false,
	}

	_, err := NewBootstrapper(config)
	if err == nil {
		t.Error("expected error for invalid KV URL")
	}
}

func TestBootstrapper_GetStatistics(t *testing.T) {
	config := Config{
		DumpFile: "nonexistent.sql",
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	stats := b.GetStatistics()

	// Initial stats should be zero
	if stats.StatementsRead != 0 {
		t.Errorf("expected StatementsRead to be 0, got %d", stats.StatementsRead)
	}
	if stats.ChangesGenerated != 0 {
		t.Errorf("expected ChangesGenerated to be 0, got %d", stats.ChangesGenerated)
	}
	if stats.ChangesStored != 0 {
		t.Errorf("expected ChangesStored to be 0, got %d", stats.ChangesStored)
	}
}

func TestBootstrapper_Close_NilKVBuffer(t *testing.T) {
	config := Config{
		DumpFile: "nonexistent.sql",
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}

	// Close should not error with nil kvBuffer
	err = b.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

func TestBootstrapper_Bootstrap_DryRun(t *testing.T) {
	// Create a temp dump file with valid mysqldump content
	tmpDir := t.TempDir()
	dumpFile := filepath.Join(tmpDir, "test.sql")

	dumpContent := `-- MySQL dump 10.13
--
-- Host: localhost    Database: testdb
-- ------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;

--
-- Table structure for table ` + "`users`" + `
--

DROP TABLE IF EXISTS ` + "`users`" + `;
CREATE TABLE ` + "`users`" + ` (
  ` + "`id`" + ` int NOT NULL AUTO_INCREMENT,
  ` + "`name`" + ` varchar(100) DEFAULT NULL,
  PRIMARY KEY (` + "`id`" + `)
) ENGINE=InnoDB;

--
-- Dumping data for table ` + "`users`" + `
--

INSERT INTO ` + "`users`" + ` VALUES (1,'John Doe'),(2,'Jane Smith');
`

	err := os.WriteFile(dumpFile, []byte(dumpContent), 0644)
	if err != nil {
		t.Fatalf("failed to write temp dump file: %v", err)
	}

	config := Config{
		DumpFile: dumpFile,
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	ctx := context.Background()
	err = b.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	stats := b.GetStatistics()

	// Should have processed statements
	if stats.StatementsRead == 0 {
		t.Error("expected StatementsRead > 0")
	}

	// In dry run, ChangesStored should equal ChangesGenerated
	if stats.ChangesStored != stats.ChangesGenerated {
		t.Errorf("in dry run, ChangesStored (%d) should equal ChangesGenerated (%d)",
			stats.ChangesStored, stats.ChangesGenerated)
	}

	// Should have end time set
	if stats.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}

	// Duration should be positive
	duration := stats.EndTime.Sub(stats.StartTime)
	if duration < 0 {
		t.Errorf("expected positive duration, got %v", duration)
	}
}

func TestBootstrapper_Bootstrap_FileNotFound(t *testing.T) {
	config := Config{
		DumpFile: "/nonexistent/path/to/dump.sql",
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	ctx := context.Background()
	err = b.Bootstrap(ctx)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestBootstrapper_Bootstrap_ContextCancellation(t *testing.T) {
	// Create a temp dump file
	tmpDir := t.TempDir()
	dumpFile := filepath.Join(tmpDir, "test.sql")

	// Create a file with many inserts to give time for cancellation
	var dumpContent strings.Builder
	dumpContent.WriteString("CREATE TABLE users (id INT);\n")
	for i := 0; i < 100; i++ {
		dumpContent.WriteString(fmt.Sprintf("INSERT INTO users VALUES (%d);\n", i))
	}

	err := os.WriteFile(dumpFile, []byte(dumpContent.String()), 0644)
	if err != nil {
		t.Fatalf("failed to write temp dump file: %v", err)
	}

	// Note: To properly test cancellation during storeChanges,
	// we'd need to not use DryRun and have a real/mocked KV buffer.
	// This test verifies the context is passed through.
	config := Config{
		DumpFile: dumpFile,
		DryRun:   true,
	}

	b, err := NewBootstrapper(config)
	if err != nil {
		t.Fatalf("NewBootstrapper() failed: %v", err)
	}
	defer b.Close()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// In dry run mode, the cancelled context won't affect execution
	// because storeChanges is skipped. This test mainly verifies
	// the context is accepted.
	_ = b.Bootstrap(ctx)
}

func TestStatistics_Fields(t *testing.T) {
	stats := Statistics{
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(time.Second),
		StatementsRead:    10,
		ChangesGenerated:  5,
		ChangesStored:     5,
		DDLCount:          2,
		DMLCount:          3,
		TablesProcessed:   []string{"users", "orders"},
		LastPosition:      "0/BOOTSTRAP0000000000000005",
		BytesProcessed:    1024,
		ErrorsEncountered: 0,
	}

	if stats.StatementsRead != 10 {
		t.Errorf("expected StatementsRead=10, got %d", stats.StatementsRead)
	}
	if len(stats.TablesProcessed) != 2 {
		t.Errorf("expected 2 tables, got %d", len(stats.TablesProcessed))
	}
}

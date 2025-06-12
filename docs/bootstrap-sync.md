# pg-bootstrap-sync Implementation Plan

## Overview

This document outlines the implementation plan for `pg-bootstrap-sync`, a CLI tool that enables bootstrapping PostgreSQL replica databases by parsing dump files and converting them to Change objects that integrate with the existing Kasho replication infrastructure.

## Architecture Goals

1. **Create shared KV buffer logic** in `pkg/kvbuffer` to avoid code duplication
2. **Build CLI tool** in `tools/pg-bootstrap-sync` for one-time bootstrap operations
3. **Maintain clean separation** between services, tools, and shared packages
4. **Reuse existing transformation logic** from pg-translicator
5. **Integrate seamlessly** with existing pg-change-stream and pg-translicator services

## Final Architecture

```
kasho/
├── pkg/                          # NEW - Internal shared packages
│   └── kvbuffer/                # NEW - Shared KV buffer logic
│       ├── go.mod
│       ├── buffer.go            # Extracted from pg-change-stream
│       └── buffer_test.go
├── tools/
│   └── pg-bootstrap-sync/        # NEW - CLI tool
│       ├── main.go
│       ├── go.mod
│       ├── internal/
│       │   ├── parser/          # Parse pg_dump files
│       │   ├── converter/       # Convert to Change objects
│       │   └── bootstrap/       # Main bootstrap logic
│       └── Dockerfile           # For containerized execution
├── services/
│   ├── pg-change-stream/        # MODIFIED - imports pkg/kvbuffer
│   └── pg-translicator/         # MODIFIED - may import pkg/kvbuffer
├── proto/                       # UNCHANGED - protobuf definitions
└── docs/                        # NEW - project documentation
```

## Implementation Steps

### Phase 1: Extract Shared KV Buffer Logic ✅

#### Step 1.1: Create pkg directory structure ✅
- [x] Create `pkg/` directory at project root
- [x] Create `pkg/kvbuffer/` subdirectory

#### Step 1.2: Create shared kvbuffer module ✅
- [x] Create `pkg/kvbuffer/go.mod` with module definition
- [x] Extract `KVBuffer` struct and methods from `services/pg-change-stream/internal/server/kvbuffer.go`
- [x] Move to `pkg/kvbuffer/buffer.go`
- [x] Extract and move unit tests to `pkg/kvbuffer/buffer_test.go`
- [x] Ensure tests pass in new location
- [x] Added type-safe `Change` interface with `Type()` and `GetLSN()` methods
- [x] Added bootstrap LSN support (format: `0/BOOTSTRAP00000001`)
- [x] Added Subscribe method for Redis pubsub integration

#### Step 1.3: Update pg-change-stream to use shared package ✅
- [x] Update `services/pg-change-stream/go.mod` to reference `pkg/kvbuffer`
- [x] Update imports in `services/pg-change-stream/internal/server/grpc.go`
- [x] Update imports in `services/pg-change-stream/cmd/server/main.go`
- [x] Remove `services/pg-change-stream/internal/server/kvbuffer.go`
- [x] Remove `services/pg-change-stream/internal/server/kvbuffer_test.go`
- [x] Run tests to ensure pg-change-stream still works
- [x] Updated `types.Change` to implement `kvbuffer.Change` interface

#### Step 1.4: Update workspace configuration ✅
- [x] Add `./pkg/kvbuffer` to `go.work`
- [x] Run `go work sync` to update workspace
- [x] Verify all modules can resolve dependencies

### Phase 2: Create pg-bootstrap-sync CLI Tool ✅

#### Step 2.1: Create tool directory structure ✅
- [x] Create `tools/pg-bootstrap-sync/` directory
- [x] Create `tools/pg-bootstrap-sync/internal/` subdirectory
- [x] Create `tools/pg-bootstrap-sync/internal/parser/` subdirectory
- [x] Create `tools/pg-bootstrap-sync/internal/converter/` subdirectory
- [x] Create `tools/pg-bootstrap-sync/internal/bootstrap/` subdirectory

#### Step 2.2: Create Go module for CLI tool ✅
- [x] Create `tools/pg-bootstrap-sync/go.mod`
- [x] Add dependencies: `pg_query_go`, `pkg/kvbuffer`, `kasho/proto`, `pglogrepl`, `cobra`
- [x] Add to `go.work` file

#### Step 2.3: Implement core parser functionality ✅
- [x] Create `tools/pg-bootstrap-sync/internal/parser/dump.go`
  - [x] Implement pg_dump file format detection
  - [x] Parse COPY statements and data
  - [x] Parse DDL statements (CREATE TABLE, etc.)
  - [x] Handle PostgreSQL COPY format escaping
  - [x] Extract table and column information
- [x] Create `tools/pg-bootstrap-sync/internal/parser/sql.go`
  - [x] Integrate `pg_query_go` for SQL parsing
  - [x] Handle complex DDL statements
  - [x] Parse CREATE TABLE, CREATE INDEX, ALTER TABLE statements
- [x] Create `tools/pg-bootstrap-sync/internal/parser/types.go`
  - [x] Define parsed statement types (DDLStatement, DMLStatement)
  - [x] Define parsing interfaces and metadata structures
- [x] Create `tools/pg-bootstrap-sync/internal/parser/dump_test.go`
  - [x] Unit tests for dump parsing functionality
  - [x] Tests for COPY statement parsing and DDL detection

#### Step 2.4: Implement Change object converter ✅
- [x] Create `tools/pg-bootstrap-sync/internal/converter/change.go`
  - [x] Convert parsed DDL to DDLData objects
  - [x] Convert parsed COPY data to DMLData objects
  - [x] Handle column type conversions (int, float, bool, string, timestamp)
  - [x] Generate proper protobuf ColumnValue objects
- [x] Create `tools/pg-bootstrap-sync/internal/converter/lsn.go`
  - [x] Implement synthetic LSN generation (`0/BOOTSTRAP00000001` format)
  - [x] Ensure LSNs are ordered and unique
  - [x] Provide LSN range validation against snapshot LSN
  - [x] Thread-safe LSN generation with mutex
- [x] Create `tools/pg-bootstrap-sync/internal/converter/types.go`
  - [x] Define conversion configuration and statistics
- [x] Create `tools/pg-bootstrap-sync/internal/converter/lsn_test.go`
  - [x] Unit tests for LSN generation and parsing
  - [x] Tests for bootstrap LSN comparison and validation
- [x] Create `tools/pg-bootstrap-sync/internal/converter/change_test.go`
  - [x] Unit tests for Change object conversion
  - [x] Tests for DDL and DML statement conversion

#### Step 2.5: Implement main bootstrap logic ✅
- [x] Create `tools/pg-bootstrap-sync/internal/bootstrap/bootstrap.go`
  - [x] Orchestrate parsing and conversion pipeline
  - [x] Handle progress tracking and statistics
  - [x] Implement error recovery and graceful shutdown
  - [x] Batch processing with configurable batch sizes
  - [x] Integration with shared kvbuffer package
- [x] Create `tools/pg-bootstrap-sync/main.go`
  - [x] CLI argument parsing using Cobra (dump file, LSN, KV URL, etc.)
  - [x] Configuration validation and setup
  - [x] Main execution flow with context cancellation
  - [x] Dry-run mode for testing
  - [x] Comprehensive help and usage information

### Phase 3: Error handling and logging ✅

- [x] Implement structured logging throughout
- [x] Add progress reporting (rows processed, time estimates)
- [x] Implement graceful shutdown handling
- [x] Add comprehensive error messages

## Technical Considerations

### LSN Strategy
- Synthetic LSNs will use format: `0/BOOTSTRAP{:08d}` (e.g., `0/BOOTSTRAP00000001`)
- All synthetic LSNs must be less than the snapshot LSN
- Bootstrap LSNs will be generated in increasing order
- LSN gaps will be left for future bootstrap operations

### Transformation Integration
- Reuse existing transformation configuration format from pg-translicator
- Apply transformations during the conversion phase, not after
- Handle transformation errors by logging and optionally skipping rows

### Memory Management
- Process dump files in streaming fashion to handle large files
- Implement configurable batch sizes for KV buffer writes
- Use checkpointing to enable resuming large operations

### Error Recovery
- Save progress checkpoints with processed LSN ranges
- Enable resuming from last checkpoint on failure
- Provide detailed error reporting with context

## Dependencies

### External Libraries
- `github.com/pganalyze/pg_query_go/v6` - PostgreSQL SQL parsing
- `github.com/redis/go-redis/v9` - Redis client (via pkg/kvbuffer)
- `github.com/spf13/cobra` - CLI framework (optional)

### Internal Dependencies
- `pkg/kvbuffer` - Shared KV buffer operations
- `kasho/proto` - Change object definitions
- `pg-translicator/internal/transform` - Transformation logic (if extracting)

## Success Criteria

1. **Functional**: pg-bootstrap-sync can parse pg_dump files and populate Redis buffer
2. **Integration**: Changes are consumable by pg-translicator without modification
3. **Performance**: Can handle multi-GB dump files without excessive memory usage
4. **Reliability**: Provides error recovery and progress tracking
5. **Maintainable**: Shared code is properly extracted and tested

## Risks and Mitigations

### Risk: pg_query_go CGO dependency
**Mitigation**: Use multi-stage Docker builds; consider fallback SQL parser

### Risk: Large dump file memory usage
**Mitigation**: Implement streaming parser with configurable batch sizes

### Risk: LSN conflicts with real WAL data
**Mitigation**: Use clear synthetic LSN format and validation

### Risk: Complex pg_dump format variations
**Mitigation**: Test with dumps from different PostgreSQL versions and configurations

## Checklist Summary

- [x] **Phase 1**: Extract shared KV buffer logic (4 steps) ✅
- [x] **Phase 2**: Create pg-bootstrap-sync CLI tool (5 steps) ✅
- [x] **Phase 3**: Error handling and logging (4 steps) ✅
- [x] **Phase 4**: Testing and validation (3 steps) ✅

**Total**: 16 steps across 4 phases  
**Completed**: 16 steps (All phases complete) ✅
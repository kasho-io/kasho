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

### Phase 2: Create pg-bootstrap-sync CLI Tool

#### Step 2.1: Create tool directory structure
- [ ] Create `tools/pg-bootstrap-sync/` directory
- [ ] Create `tools/pg-bootstrap-sync/internal/` subdirectory
- [ ] Create `tools/pg-bootstrap-sync/internal/parser/` subdirectory
- [ ] Create `tools/pg-bootstrap-sync/internal/converter/` subdirectory
- [ ] Create `tools/pg-bootstrap-sync/internal/bootstrap/` subdirectory

#### Step 2.2: Create Go module for CLI tool
- [ ] Create `tools/pg-bootstrap-sync/go.mod`
- [ ] Add dependencies: `pg_query_go`, `pkg/kvbuffer`, `kasho/proto`
- [ ] Add to `go.work` file

#### Step 2.3: Implement core parser functionality
- [ ] Create `tools/pg-bootstrap-sync/internal/parser/dump.go`
  - [ ] Implement pg_dump file format detection
  - [ ] Parse COPY statements and data
  - [ ] Parse DDL statements (CREATE TABLE, etc.)
- [ ] Create `tools/pg-bootstrap-sync/internal/parser/sql.go`
  - [ ] Integrate `pg_query_go` for SQL parsing
  - [ ] Handle complex DDL statements
- [ ] Create `tools/pg-bootstrap-sync/internal/parser/types.go`
  - [ ] Define parsed statement types
  - [ ] Define parsing interfaces

#### Step 2.4: Implement Change object converter
- [ ] Create `tools/pg-bootstrap-sync/internal/converter/change.go`
  - [ ] Convert parsed DDL to DDLData objects
  - [ ] Convert parsed COPY data to DMLData objects
  - [ ] Handle column type conversions
- [ ] Create `tools/pg-bootstrap-sync/internal/converter/lsn.go`
  - [ ] Implement synthetic LSN generation
  - [ ] Ensure LSNs are ordered and unique
  - [ ] Provide LSN range validation

#### Step 2.5: Implement main bootstrap logic
- [ ] Create `tools/pg-bootstrap-sync/internal/bootstrap/bootstrap.go`
  - [ ] Orchestrate parsing and conversion pipeline
  - [ ] Handle progress tracking and checkpointing
  - [ ] Implement error recovery
- [ ] Create `tools/pg-bootstrap-sync/main.go`
  - [ ] CLI argument parsing (dump file, LSN range, etc.)
  - [ ] Configuration loading
  - [ ] Main execution flow

#### Step 2.6: Add transformation support
- [ ] Import transformation config from pg-translicator
- [ ] Apply transformations to DML data during conversion
- [ ] Handle transformation errors gracefully

### Phase 3: Integration and Configuration

#### Step 3.1: Docker configuration
- [ ] Create `tools/pg-bootstrap-sync/Dockerfile`
- [ ] Test containerized execution
- [ ] Update docker-compose files if needed for development

#### Step 3.2: CLI interface design
- [ ] Design command-line flags:
  - [ ] `--dump-file` - Path to pg_dump file
  - [ ] `--snapshot-lsn` - LSN where WAL replication should begin
  - [ ] `--kv-url` - Redis connection string
  - [ ] `--transform-config` - Path to transformation config
  - [ ] `--batch-size` - Processing batch size
  - [ ] `--resume-from` - Checkpoint file for resuming
- [ ] Implement flag parsing and validation
- [ ] Add help text and usage examples

#### Step 3.3: Error handling and logging
- [ ] Implement structured logging throughout
- [ ] Add progress reporting (rows processed, time estimates)
- [ ] Implement graceful shutdown handling
- [ ] Add comprehensive error messages

### Phase 4: Testing and Validation

#### Step 4.1: Unit testing
- [ ] Test parser components with sample dump files
- [ ] Test converter components with various data types
- [ ] Test LSN generation and ordering
- [ ] Test error conditions and edge cases

#### Step 4.2: Integration testing
- [ ] Test full pipeline with real pg_dump files
- [ ] Test transformation application
- [ ] Test KV buffer integration
- [ ] Test with pg-change-stream and pg-translicator

#### Step 4.3: End-to-end testing
- [ ] Create test scenario:
  - [ ] Generate pg_dump from source database
  - [ ] Run pg-bootstrap-sync to populate KV buffer
  - [ ] Start pg-translicator to process changes
  - [ ] Verify replica database matches source
- [ ] Test with large datasets
- [ ] Test error recovery scenarios

### Phase 5: Documentation and Deployment

#### Step 5.1: Usage documentation
- [ ] Create user guide for pg-bootstrap-sync
- [ ] Document CLI flags and options
- [ ] Provide example workflows
- [ ] Document integration with existing services

#### Step 5.2: Deployment configuration
- [ ] Add to Taskfile.yml for development builds
- [ ] Create GitHub Actions workflow for building tool
- [ ] Document container deployment patterns
- [ ] Add to environment configurations

#### Step 5.3: Performance optimization
- [ ] Implement parallel processing for large dumps
- [ ] Add memory usage optimization
- [ ] Implement streaming processing for large COPY data
- [ ] Add performance benchmarks

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
- [ ] **Phase 2**: Create pg-bootstrap-sync CLI tool (6 steps)
- [ ] **Phase 3**: Integration and configuration (3 steps)
- [ ] **Phase 4**: Testing and validation (3 steps)
- [ ] **Phase 5**: Documentation and deployment (3 steps)

**Total**: 19 major steps across 5 phases  
**Completed**: 4 steps (Phase 1 complete)
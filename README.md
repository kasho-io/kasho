# Translicate

Translicate aims to be a PostgreSQL replication tool that captures and applies DDL (Data Definition Language) and DML (Data Manipulation Language) changes from a primary database. For DML changes, each will be fed through a transform layer in order to modify / sanitize things so that no sensitive data is exposed to the replica database. It uses PostgreSQL's logical replication capabilities to ensure that schema changes and data modifications are properly synchronized.

## Features

- Captures DDL changes using a custom trigger-based logging system
- Applies DDL changes in the correct order before processing DML changes
- Handles both new and existing DDL changes when starting replication
- Uses PostgreSQL's native logical replication for reliable change capture
- Supports table creation, column additions, and other schema modifications

## Plans

- Allow complex transforms on data
  - Possibly use an LLM to help figure out what needs to be transformed
- Support bootstrapping from an existing database with significant schema and data

## Components

### pg-change-stream (`services/pg-change-stream/`)
- Go service that captures and streams database changes
- Uses PostgreSQL's logical replication
- Provides SSE interface for real-time change streaming
- Handles connection management and retries

### replicate-poc (`services/replicate-poc`)
- Connects to both primary and replica databases
- Polls the `translicate_ddl_log` table for DDL changes
- Applies DDL changes in order before processing DML changes
- Uses PostgreSQL's logical replication to capture and apply DML changes
- Handles both startup scenarios:
  1. Starting before changes: Captures and applies changes as they occur
  2. Starting after changes: Applies historical DDLs before processing DMLs

### Development Environment (`environments/development/`)
- Docker Compose setup for local development
- Includes PostgreSQL primary and replica databases
- Redis for caching
- pg-change-stream service

## Getting Started

1. Install dependencies:
   - Docker
   - Go 1.24 or later
   - Task (task runner)

2. Install Task:
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

3. Reset and restart the development environment:
```bash
task dev-reset
```

4. Run tests:
```bash
task test
```

## Development

The project uses Task for common development commands. Available commands:

- `task dev`: Start development environment
- `task dev-down`: Stop development environment
- `task dev-reset`: Reset and restart development environment (removes volumes)
- `task build`: Build the pg-change-stream service
- `task test`: Run tests for pg-change-stream service

## Requirements

- Docker
- Go 1.24 or later
- Task (task runner)
- Python 3.13 or later (for replicate-poc)

## License

Copyright &copy; Jeffrey Wescott, 2025. All rights reserved.
 
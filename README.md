# Kasho

Kasho is a security-and-privacy-first PostgreSQL replication tool that captures and applies DDL (Data Definition Language) and DML (Data Manipulation Language) changes from a primary database to a replica database. For DML changes, each will be fed through a transform layer in order to modify / sanitize things so that no sensitive data is exposed to the replica database. It uses PostgreSQL's logical replication capabilities to ensure that schema changes and data modifications are properly synchronized.

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

### Services

#### pg-change-stream (`services/pg-change-stream/`)
- Go service that captures and streams database changes
- Uses PostgreSQL's logical replication
- Provides gRPC interface for real-time change streaming
- Handles connection management and retries
- Uses Redis for buffering and pub/sub

#### pg-translicator (`services/pg-translicator/`)
- Go service for translating and transforming database changes
- Processes change events from pg-change-stream
- Supports custom transformation rules
- Integrates with external systems

#### replicate-poc (`services/replicate-poc/`)
- Proof of concept for replication features
- Experimental service for testing new replication patterns
- Used for validating replication strategies

### Environments

#### Development (`environments/development/`)
- Docker Compose setup for local development
- Includes PostgreSQL primary and replica databases
- Redis for caching
- All core services configured for development

#### Demo (`environments/demo/`)
- Demo environment with PostgreSQL primary and replica databases
- Configured for logical replication
- Used for testing replication features
- Includes initialization scripts for database setup

#### POC (`environments/poc/`)
- Environment for proof of concept testing
- Used for validating new features
- Includes experimental configurations

### Tools (`tools/`)
- `generate-fake-saas-data`: Utility for generating realistic SaaS application test data
  - Creates sample organizations, users, subscriptions, and related data
  - Used for populating test databases with realistic data

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

The project uses Task for common development commands. Type `task` by itself for a list of commands.

## Requirements

- Docker
- Go 1.24 or later
- Task (task runner)
- Python 3.13 or later (for replicate-poc)

## License

Copyright &copy; Jeffrey Wescott, 2025. All rights reserved.
 
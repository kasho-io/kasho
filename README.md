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

### Environments

#### Development (`environments/development/`)
- Docker Compose setup for local development
- Includes PostgreSQL primary and replica databases
- Redis for caching
- All core services configured for development

### Tools (`tools/`)
- `generate-fake-saas-data`: Utility for generating realistic SaaS application test data
  - Creates sample organizations, users, subscriptions, and related data
  - Used for populating test databases with realistic data

### SQL (`sql/`)
- Scripts that will be needed to get a Postgres server ready to use kasho
- They should be run in the order that they are numbered
- `00-setup-wal-level.sql` is not necessary in development because the `docker-compose.yml` file handles it.

### Apps

#### Homepage (`apps/homepage/`)
- Public-facing landing page for Kasho.
- Shows the Kasho wordmark and a brief description.
- To run locally:
  ```bash
  task homepage
  ```
  (Runs on [http://localhost:3000](http://localhost:3000))

#### Demo (`apps/demo/`)
- Interactive demo app for showcasing Kasho's real-time database replication and transformation features.
- Includes a UI for connecting to primary and replica databases, and visualizing real-time changes.
- To run locally:
  ```bash
  task demo
  ```
  (Runs on [http://localhost:4000](http://localhost:4000))

## Development

The project uses Docker, Golang, and Task.

1. Install dependencies:
   - Docker
   - Go 1.24 or later
   - Task (task runner)

2. Install Task:
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

Type `task` by itself for a list of commands.

## Getting Started

1. Reset and start the development environment. In a dedicated terminal, run:
```bash
task dev-reset
```

2. Setup replication and some test data. In a separate terminal, run:
```bash
task dev-setup
```

3. Run the demo. In a separate terminal, run:
```bash
task demo
```

## Requirements

- Docker
- Go 1.24 or later
- Task (task runner)
- Python 3.13 or later (for replicate-poc)

## License

Copyright &copy; Jeffrey Wescott, 2025. All rights reserved.
 
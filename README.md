# Kasho

Kasho is a security-and-privacy-first PostgreSQL replication tool that captures and applies DDL (Data Definition Language) and DML (Data Manipulation Language) changes from a primary database to a replica database. For DML changes, each will be fed through a transform layer in order to modify / sanitize things so that no sensitive data is exposed to the replica database. It uses PostgreSQL's logical replication capabilities to ensure that schema changes and data modifications are properly synchronized.

## Features

- Captures DDL changes using a custom trigger-based logging system
- Applies DDL changes in the correct order before processing DML changes
- Handles both new and existing DDL changes when starting replication
- Uses PostgreSQL's native logical replication for reliable change capture
- Supports table creation, column additions, and other schema modifications

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
- All core services configured for development with hot reload

#### Demo (`environments/demo/`)
- Production-like environment for demonstration and testing
- Uses production Docker image builds
- Includes PostgreSQL primary and replica databases
- All core services configured for production deployment
- Automated deployment via GitHub Actions

### Tools (`tools/`)
- `pg-bootstrap-sync`: CLI tool for bootstrapping replica databases from PostgreSQL dump files
  - Parses pg_dump files and converts them to Change objects for the replication system
  - Supports both COPY and INSERT format dump files
  - Integrates with the shared kvbuffer for seamless data flow
- `env-template`: Environment variable substitution utility for configuration templates
- `generate-fake-saas-data`: Utility for generating realistic SaaS application test data
  - Creates sample organizations, users, subscriptions, and related data
  - Used for populating test databases with realistic data

### Shared Packages (`pkg/`)
- `kvbuffer`: Shared Redis-based buffer for change events
  - Used by both pg-change-stream and pg-bootstrap-sync
  - Supports both real PostgreSQL LSNs and synthetic bootstrap LSNs
  - Provides type-safe Change interface
- `types`: Common type definitions and utilities
  - JSON marshaling wrappers for protobuf types
  - Shared data structures across services
  - Type conversion utilities for database change events

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
  task dev:app-homepage
  ```
  (Runs on [http://localhost:3000](http://localhost:3000))

#### Demo (`apps/demo/`)
- Interactive demo app for showcasing Kasho's real-time database replication and transformation features.
- Includes a UI for connecting to primary and replica databases, and visualizing real-time changes.
- To run locally:
  ```bash
  task dev:app-demo
  ```
  (Runs on [http://localhost:4000](http://localhost:4000))

## Architecture

Kasho uses a consolidated Docker approach for simplified deployment:

- **Single Dockerfile** at the root builds all services and tools into one image
- **Multi-stage builds** with development and production targets
- **Environment-specific configurations** in `environments/` directory
- **Task-based development** workflow for local development and testing

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

## Deployment

Kasho is packaged as a single container image containing all services and tools:

```bash
# Build the consolidated image
docker build -t kasho .

# Run individual services
docker run kasho ./pg-change-stream
docker run kasho ./pg-translicator  
docker run kasho ./pg-bootstrap-sync --help

# Or use environment-specific docker-compose files
cd environments/demo && docker-compose up
```

## Getting Started

1. Reset and start the development environment. In a dedicated terminal, run:
```bash
task dev:reset
```

2. Setup replication. In a separate terminal, run:
```bash
task dev:setup-replication
```

3. Wait for the `pg-change-stream` service to connect to the primary by watching the output from the first terminal, then set up some test data. Run:
```bash
task dev:setup-data
```

4. Run the demo. In a separate terminal, run:
```bash
task dev:app-demo
```

## Requirements

- Docker
- Go 1.24 or later
- Task (task runner)

## License

Copyright &copy; Jeffrey Wescott, 2025. All rights reserved.
 
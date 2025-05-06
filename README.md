# Translicate

Translicate is a PostgreSQL replication tool that captures and applies DDL (Data Definition Language) and DML (Data Manipulation Language) changes from a primary database to a replica database. It uses PostgreSQL's logical replication capabilities to ensure that schema changes and data modifications are properly synchronized.

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
- Port to Golang (and possibly Temporal.io)

## Components

### SQL Setup Scripts (`sql/`)
- `setup_wal_primary.sql`: Enables WAL logging on the primary database
- `setup_replication_primary.sql`: Sets up the primary database with DDL logging and replication
- `setup_replication_replica.sql`: Sets up the replica database for replication
- Creates the DDL logging system with triggers and necessary permissions

### Python Consumer (`python-consumer/translicate_consumer.py`)
- Connects to both primary and replica databases
- Polls the `translicate_ddl_log` table for DDL changes
- Applies DDL changes in order before processing DML changes
- Uses PostgreSQL's logical replication to capture and apply DML changes
- Handles both startup scenarios:
  1. Starting before changes: Captures and applies changes as they occur
  2. Starting after changes: Applies historical DDLs before processing DMLs

### Demo using Docker (`demo/`)

- `Dockerfile`: build the PostgreSQL image
- `docker-compose.yml`: start two containers, one primary and one replica, for PostgreSQL
- `scripts`: scripts for running the demo
- `sql`: sql files for running the demo

## Getting Started

1. Install Docker.

2. Start the demo.
```bash
cd demo/scripts
docker-compose up
```

3. Reset the databases.
```bash
./demo/scripts/reset-dbs.sh
```

4. Set up replication.
```bash
./demo/scripts/setup-replication.sh
```

5. Run the demo consumer.
```bash
cd python-consumer
pipenv install
cat .env.demo > .env
python3 translicate_consumer.py
```

6. Execute some DB operations
```bash
./demo/scripts/demo_replication.sh
```

7. Make sure there are no differences across the primary and replica.
```bash
./demo/scripts/compare-dbs.sh
```
## Requirements

- Docker
- Python 3.13 or later

## License

Copyright &copy; Jeffrey Wescott, 2025. All rights reserved.
 
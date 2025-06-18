# Kasho Deployment Guide

Kasho is a data synchronization platform that consists of multiple cooperating services. While it has multiple components under the hood, we've designed deployment to be straightforward.

## Architecture Overview

Kasho consists of three main components:

1. **pg-change-stream**: Captures changes from PostgreSQL using logical replication
2. **pg-translicator**: Applies changes to the target database  
3. **Redis**: Provides message buffering between services

```
PostgreSQL Source → pg-change-stream → Redis → pg-translicator → Target Database
```

**Important:** Currently, only a single pg-translicator instance is supported. Multiple translicators would process the same changes multiple times.

## Configuration Requirements

### Environment Variables

#### pg-change-stream
| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `KV_URL` | Redis connection URL | Yes | `redis://redis:6379` |
| `PRIMARY_DATABASE_URL` | Source database connection URL | Yes | `postgresql://user:pass@host:5432/db?sslmode=disable` |

#### pg-translicator
| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `CHANGE_STREAM_SERVICE` | pg-change-stream service address | Yes | `pg-change-stream:8080` |
| `REPLICA_DATABASE_URL` | Target database connection URL | Yes | `postgresql://user:pass@host:5432/db?sslmode=disable` |

#### URL Format
Database URLs follow the standard PostgreSQL connection string format:
```
postgresql://[user[:password]@][host][:port][/database][?param=value&...]
```

Common parameters:
- `sslmode=disable|require|verify-ca|verify-full`
- `connect_timeout=30`
- `application_name=kasho`


### Configuration Files

pg-translicator requires a `transforms.yml` file mounted at `/app/config/transforms.yml`. This file defines table and column transformations.

For detailed information about the transforms.yml file format, available transform types, and configuration examples, see the [Transform Configuration Guide](transforms.md).

## Deployment

Since production deployment environments differ so significantly across different organizations, it's difficult to give perfect instructions for every environment. That said, in order to illustrate how to set up and use Kasho, here's how it could be done with Docker Compose.

### 1. Docker Compose Deployment

Create a `docker-compose.yml`:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  pg-change-stream:
    image: kasho:latest
    command: ./pg-change-stream
    environment:
      KV_URL: redis://redis:6379
      PRIMARY_DATABASE_URL: ${PRIMARY_DATABASE_URL}
    ports:
      - "8080:8080"
    depends_on:
      - redis

  pg-translicator:
    image: kasho:latest
    command: ./pg-translicator
    environment:
      CHANGE_STREAM_SERVICE: pg-change-stream:8080
      REPLICA_DATABASE_URL: ${REPLICA_DATABASE_URL}
    volumes:
      - ./config:/app/config:ro
    depends_on:
      - pg-change-stream

volumes:
  redis-data:
```

**Quick Start:**
```bash
# Create config directory with transforms.yml
mkdir -p config
cp environments/demo/config/transforms.yml config/

# Set environment variables in .env file
# Start all services
docker-compose up -d
```

## Database Requirements

**Source PostgreSQL:**
- Version 10+
- Logical replication enabled (`wal_level = logical`)
- User with REPLICATION permission

**Target Database:**
- PostgreSQL 10+
- User with table creation and write permissions

## Troubleshooting

### Services Won't Start

1. Check environment variables are set correctly
2. Verify database connectivity
3. Ensure transforms.yml is mounted for pg-translicator

### Connection Issues

Check that services can reach each other:
- pg-change-stream must be accessible on port 8080
- Redis must be accessible on its configured port
- Both databases must be reachable from their respective services

### Common Errors

**"PRIMARY_DATABASE_URL environment variable is required"**
- Set PRIMARY_DATABASE_URL in the format: `postgresql://user:password@host:port/database?sslmode=disable`
- Use `docker run --rm kasho` to see the help text with examples

**"Required config file /app/config/transforms.yml not found"**
- Mount the config directory with transforms.yml to /app/config

**"replication slot already exists"**
- Only one pg-change-stream instance can use a replication slot
- Drop the existing slot or use a different slot name
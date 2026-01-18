#!/bin/sh
# kasho-help.sh - Help script for Kasho Docker image

cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                              KASHO                                ║
║              Database Change Data Capture & Replication           ║
╚═══════════════════════════════════════════════════════════════════╝

SERVICES:
  PostgreSQL:
    /app/bin/pg-change-stream       - PostgreSQL change stream service
    /app/bin/pg-bootstrap-sync      - Bootstrap replica from pg_dump

  MySQL:
    /app/bin/mysql-change-stream    - MySQL change stream service
    /app/bin/mysql-bootstrap-sync   - Bootstrap replica from mysqldump

  Shared:
    /app/bin/translicator           - Transform and apply changes to replica

SCRIPTS:
  /app/scripts/prepare-primary-db.sh - Setup primary database for Kasho replication
  /app/scripts/bootstrap-kasho.sh    - Bootstrap Kasho replication from existing data

EXAMPLES:
  # Start PostgreSQL change stream service
  docker run --rm kasho /app/bin/pg-change-stream

  # Start MySQL change stream service
  docker run --rm kasho /app/bin/mysql-change-stream

  # Start translicator service
  docker run --rm kasho /app/bin/translicator

  # Bootstrap Kasho replication (from inside running container)
  docker exec -it <container-name> /app/scripts/bootstrap-kasho.sh

  # Bootstrap using dedicated container (PostgreSQL)
  docker run --rm --network <network-name> \
    -e PRIMARY_DATABASE_URL="postgresql://..." \
    -e KV_URL="redis://..." \
    -e CHANGE_STREAM_SERVICE_ADDR="pg-change-stream:50051" \
    kasho /app/scripts/bootstrap-kasho.sh

BOOTSTRAP WORKFLOW:
  1. Prepare primary database (run as DBA with superuser credentials):
     docker exec -it <container> /app/scripts/prepare-primary-db.sh

  2. Start services (change-stream will be in WAITING state).

  3. Bootstrap replication:
     docker exec -it <container> /app/scripts/bootstrap-kasho.sh

ENVIRONMENT VARIABLES:

pg-change-stream service:
  KV_URL                 - Redis connection URL
                          Example: redis://127.0.0.1:6379

  PRIMARY_DATABASE_URL   - Primary database connection URL
                          Example: postgresql://user:pass@host:5432/db?sslmode=disable

mysql-change-stream service:
  KV_URL                 - Redis connection URL
                          Example: redis://127.0.0.1:6379

  PRIMARY_DATABASE_URL   - Primary database connection URL
                          Example: mysql://user:pass@host:3306/db

translicator service:
  CHANGE_STREAM_SERVICE_ADDR - Change stream service address
                              Example: pg-change-stream:50051 or mysql-change-stream:50051

  REPLICA_DATABASE_URL       - Replica database connection URL
                              Example: postgresql://user:pass@host:5432/db?sslmode=disable
                              Example: mysql://user:pass@host:3306/db

  DATABASE_TYPE              - Database type (postgresql or mysql)
                              Default: postgresql

URL FORMATS:
  PostgreSQL: postgresql://[user[:password]@][host][:port][/database][?param=value&...]
  MySQL:      mysql://[user[:password]@][host][:port][/database][?param=value&...]

  Common PostgreSQL parameters:
    sslmode=disable|require|verify-ca|verify-full
    connect_timeout=30
    application_name=myapp

DOCUMENTATION:
  For full documentation, visit: https://docs.kasho.io

  Configuration files are located in:
    /app/config/demo/      - Demo environment configuration
    /app/config/development/ - Development environment configuration

TROUBLESHOOTING:
  • Ensure databases are reachable before starting services
  • Check that KV_URL points to a running Redis instance
  • Verify SSL settings match your database configuration
  • Use docker logs <container> to view service output

EOF

#!/bin/sh
# kasho-help.sh - Help script for Kasho Docker image

cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                              KASHO                                ║
║                PostgreSQL Change Data Capture                    ║
╚═══════════════════════════════════════════════════════════════════╝

SERVICES:
  ./pg-change-stream     - Start change stream service
  ./pg-translicator      - Start translicator service

TOOLS:
  ./pg-bootstrap-sync    - Bootstrap replica databases from pg_dump
  ./env-template         - Process environment variable templates

EXAMPLES:
  # Start change stream service
  docker run --rm kasho ./pg-change-stream

  # Start translicator service  
  docker run --rm kasho ./pg-translicator

  # Bootstrap a database
  docker run --rm -v ./dump.sql:/dump.sql kasho \
    ./pg-bootstrap-sync --source /dump.sql --kv-url redis://host:6379

ENVIRONMENT VARIABLES:

pg-change-stream service:
  KV_URL                 - Redis connection URL
                          Example: redis://127.0.0.1:6379
  
  PRIMARY_DATABASE_URL   - Primary database connection URL
                          Example: postgresql://user:pass@host:5432/db?sslmode=disable

pg-translicator service:
  CHANGE_STREAM_SERVICE  - Change stream service address
                          Example: pg-change-stream:8080
  
  REPLICA_DATABASE_URL   - Replica database connection URL
                          Example: postgresql://user:pass@host:5432/db?sslmode=disable

URL FORMATS:
  Database URLs: postgresql://[user[:password]@][host][:port][/database][?param=value&...]
  
  Common parameters:
    sslmode=disable|require|verify-ca|verify-full
    connect_timeout=30
    application_name=myapp

DOCUMENTATION:
  For full documentation, visit: https://github.com/kasho/kasho
  
  Configuration files are located in:
    /app/config/demo/      - Demo environment configuration
    /app/config/development/ - Development environment configuration

TROUBLESHOOTING:
  • Ensure databases are reachable before starting services
  • Check that KV_URL points to a running Redis instance
  • Verify SSL settings match your database configuration
  • Use docker logs <container> to view service output

EOF
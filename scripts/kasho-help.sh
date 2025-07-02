#!/bin/sh
# kasho-help.sh - Help script for Kasho Docker image

cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                              KASHO                                ║
║                PostgreSQL Change Data Capture                    ║
╚═══════════════════════════════════════════════════════════════════╝

SERVICES:
  /app/bin/pg-change-stream     - Start change stream service
  /app/bin/pg-translicator      - Start translicator service
  /app/bin/licensing            - Start licensing service

SCRIPTS:
  /app/scripts/prepare-primary-db.sh - Setup primary database for Kasho replication
  /app/scripts/bootstrap-kasho.sh    - Bootstrap Kasho replication from existing data

EXAMPLES:
  # Start change stream service
  docker run --rm kasho /app/bin/pg-change-stream

  # Start translicator service  
  docker run --rm kasho /app/bin/pg-translicator

  # Bootstrap Kasho replication (from inside running container)
  docker exec -it <container-name> /app/scripts/bootstrap-kasho.sh

  # Bootstrap using dedicated container
  docker run --rm --network <network-name> \
    -e PRIMARY_DATABASE_URL="postgresql://..." \
    -e KV_URL="redis://..." \
    -e CHANGE_STREAM_SERVICE_ADDR="pg-change-stream:50051" \
    kasho /app/scripts/bootstrap-kasho.sh

BOOTSTRAP WORKFLOW:
  1. Prepare primary database (run as DBA with superuser credentials):
     docker exec -it <container> /app/scripts/prepare-primary-db.sh
  
  2. Start services (pg-change-stream will be in WAITING state).
  
  3. Bootstrap replication:
     docker exec -it <container> /app/scripts/bootstrap-kasho.sh

ENVIRONMENT VARIABLES:

pg-change-stream service:
  KV_URL                 - Redis connection URL
                          Example: redis://127.0.0.1:6379
  
  PRIMARY_DATABASE_URL   - Primary database connection URL
                          Example: postgresql://user:pass@host:5432/db?sslmode=disable

pg-translicator service:
  CHANGE_STREAM_SERVICE_ADDR - Change stream service address
                              Example: pg-change-stream:50051
  
  REPLICA_DATABASE_URL       - Replica database connection URL
                              Example: postgresql://user:pass@host:5432/db?sslmode=disable

All services:
  LICENSING_SERVICE_ADDR     - License service address
                              Example: licensing:50052

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
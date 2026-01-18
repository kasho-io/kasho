#!/bin/sh
# parse-db-url.sh - Parse database URLs and export decomposed variables for env-template
# Usage: eval $(./parse-db-url.sh)

set -e

# Function to parse a database URL and export variables with a prefix
parse_db_url() {
    local url="$1"
    local prefix="$2"

    if [ -z "$url" ]; then
        return 0
    fi

    # Extract components using trurl
    local scheme=$(trurl --url "$url" --get '{scheme}' 2>/dev/null || echo "")
    local user=$(trurl --url "$url" --get '{user}' 2>/dev/null || echo "")
    local password=$(trurl --url "$url" --get '{password}' 2>/dev/null || echo "")
    local host=$(trurl --url "$url" --get '{host}' 2>/dev/null || echo "")
    local port=$(trurl --url "$url" --get '{port}' 2>/dev/null || echo "")
    local dbname=$(trurl --url "$url" --get '{path}' 2>/dev/null | sed 's|^/||' || echo "")
    local sslmode=$(trurl --url "$url" --get '{query:sslmode}' 2>/dev/null || echo "")

    # Set default port based on scheme if not specified
    if [ -z "$port" ]; then
        case "$scheme" in
            mysql)
                port="3306"
                ;;
            postgresql|postgres)
                port="5432"
                ;;
            *)
                port="5432"
                ;;
        esac
    fi
    
    # Export variables with prefix
    echo "export ${prefix}_KASHO_USER='$user'"
    echo "export ${prefix}_KASHO_PASSWORD='$password'"
    echo "export ${prefix}_HOST='$host'"
    echo "export ${prefix}_PORT='$port'"
    echo "export ${prefix}_DB='$dbname'"
    if [ -n "$sslmode" ]; then
        echo "export ${prefix}_SSLMODE='$sslmode'"
    fi
}

# Check if trurl is available
if ! command -v trurl >/dev/null 2>&1; then
    echo "Error: trurl command not found. Please install trurl." >&2
    exit 1
fi

# Parse PRIMARY_DATABASE_URL if set
if [ -n "${PRIMARY_DATABASE_URL:-}" ]; then
    parse_db_url "$PRIMARY_DATABASE_URL" "PRIMARY_DATABASE"
fi

# Parse REPLICA_DATABASE_URL if set
if [ -n "${REPLICA_DATABASE_URL:-}" ]; then
    parse_db_url "$REPLICA_DATABASE_URL" "REPLICA_DATABASE"
fi

# Set superuser variables if not already set (for backward compatibility)
if [ -n "${PRIMARY_DATABASE_URL:-}" ] && [ -z "${PRIMARY_DATABASE_SU_USER:-}" ]; then
    echo "export PRIMARY_DATABASE_SU_USER='postgres'"
    echo "export PRIMARY_DATABASE_SU_PASSWORD='password'"
fi

if [ -n "${REPLICA_DATABASE_URL:-}" ] && [ -z "${REPLICA_DATABASE_SU_USER:-}" ]; then
    echo "export REPLICA_DATABASE_SU_USER='postgres'"
    echo "export REPLICA_DATABASE_SU_PASSWORD='password'"
fi
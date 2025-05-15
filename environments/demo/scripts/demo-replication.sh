#!/bin/bash

# Exit on any error
set -e

# Check if environment argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <environment>"
    echo "Environment must be one of: poc, demo, development"
    exit 1
fi

ENV=$1
if [[ ! "$ENV" =~ ^(poc|demo|development)$ ]]; then
    echo "Error: Environment must be one of: poc, demo, development"
    exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../sql"

echo "Setting up users table and data..."

# Step 1: Create users table
echo "Creating users table..."
cat "$SQL_DIR/create_users_table.sql" | docker exec -i ${ENV}-postgres-primary-1 psql -U postgres -d source_db

# Step 2: Insert 4 users
echo "Inserting 4 users..."
for i in {1..4}; do
    echo "Inserting user $i..."
    cat "$SQL_DIR/insert_user.sql" | docker exec -i ${ENV}-postgres-primary-1 psql -U postgres -d source_db
done

# Step 3: Add DOB column
echo "Adding DOB column..."
cat "$SQL_DIR/add_dob_column_to_users_table.sql" | docker exec -i ${ENV}-postgres-primary-1 psql -U postgres -d source_db

# Step 4: Update DOB values
echo "Updating DOB values..."
cat "$SQL_DIR/update_dob_for_all_users.sql" | docker exec -i ${ENV}-postgres-primary-1 psql -U postgres -d source_db

echo "Users setup complete!" 
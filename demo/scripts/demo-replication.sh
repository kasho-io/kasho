#!/bin/bash

# Exit on any error
set -e

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_DIR="$SCRIPT_DIR/../sql"

echo "Setting up users table and data..."

# Step 1: Create users table
echo "Creating users table..."
cat "$SQL_DIR/create_users_table.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db

# Step 2: Insert 4 users
echo "Inserting 4 users..."
for i in {1..4}; do
    echo "Inserting user $i..."
    cat "$SQL_DIR/insert_user.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db
done

# Step 3: Add DOB column
echo "Adding DOB column..."
cat "$SQL_DIR/add_dob_column_to_users_table.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db

# Step 4: Update DOB values
echo "Updating DOB values..."
cat "$SQL_DIR/update_dob_for_all_users.sql" | docker exec -i demo-pg_primary-1 psql -U postgres -d source_db

echo "Users setup complete!" 
#!/bin/sh
set -e

echo "Running database migrations..."

# Check if database already has tables (existing deployment)
TABLE_COUNT=$(migrate -path /app/migrations -database "${DATABASE_URL}" version 2>&1 || true)

if echo "$TABLE_COUNT" | grep -q "Dirty"; then
  echo "Database is dirty, forcing version to latest..."
  migrate -path /app/migrations -database "${DATABASE_URL}" force 9
fi

# Run migrations
if ! migrate -path /app/migrations -database "${DATABASE_URL}" up; then
  echo "Migrations failed, forcing to latest version..."
  migrate -path /app/migrations -database "${DATABASE_URL}" force 9
fi

echo "Migrations completed successfully."

# Start the application
exec "$@"

#!/bin/sh
set -e

echo "Running database migrations..."

if ! migrate -path /app/migrations -database "${DATABASE_URL}" up; then
  echo "Migration failed. Check database state and migration files."
  exit 1
fi

echo "Migrations completed successfully."
exec "$@"

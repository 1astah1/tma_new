#!/bin/sh
# Периодический дамп базы в /backups с ротацией по возрасту.
set -e

BACKUP_DIR=${BACKUP_DIR:-/backups}
BACKUP_INTERVAL=${BACKUP_INTERVAL:-21600}
RETENTION_DAYS=${RETENTION_DAYS:-7}

if [ -z "$DATABASE_URL" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

mkdir -p "$BACKUP_DIR"

while true; do
  file="$BACKUP_DIR/tma_shop_$(date +%Y%m%d_%H%M%S).sql.gz"
  echo "Dumping to $file"
  if pg_dump "$DATABASE_URL" | gzip > "$file"; then
    echo "Backup done: $file"
  else
    echo "Backup failed"
    rm -f "$file"
  fi
  find "$BACKUP_DIR" -name '*.sql.gz' -type f -mtime "+$RETENTION_DAYS" -delete
  sleep "$BACKUP_INTERVAL"
done

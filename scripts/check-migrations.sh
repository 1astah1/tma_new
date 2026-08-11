#!/bin/sh
# Проверка каталога миграций под golang-migrate:
# каждая версия должна встречаться ровно один раз в up и один раз в down,
# иначе migrate падает с duplicate migration version.
set -e

DIR="$(dirname "$0")/../tma-backend/migrations"
status=0

for direction in up down; do
  versions=$(find "$DIR" -name "*.$direction.sql" -exec basename {} \; | cut -d_ -f1 | sort)
  dupes=$(echo "$versions" | uniq -d)
  if [ -n "$dupes" ]; then
    echo "Дубли версий ($direction): $dupes"
    status=1
  fi
done

for f in "$DIR"/*.up.sql; do
  down="${f%.up.sql}.down.sql"
  [ -f "$down" ] || { echo "Нет down-миграции для $(basename "$f")"; status=1; }
done

for f in "$DIR"/*.sql; do
  case "$(basename "$f")" in
    *.up.sql|*.down.sql) ;;
    *) echo "Файл без направления up/down: $(basename "$f")"; status=1 ;;
  esac
done

[ "$status" -eq 0 ] && echo "Миграции в порядке"
exit "$status"

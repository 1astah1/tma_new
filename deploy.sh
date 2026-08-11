#!/bin/sh
# Выкат на боевой сервер.
#
# ВАЖНО: --delete нужен, иначе удалённые файлы остаются на сервере и ломают
# сборку. Но он же однажды снёс /opt/tma/.env и каталог с бэкапами — их нет
# в репозитории. Поэтому всё, что живёт только на сервере, защищено --exclude.
set -e

HOST=${1:-root@51.38.154.221}
TARGET=/opt/tma

# Пароль сервера берётся из окружения: SSHPASS=... ./deploy.sh
SSH="ssh"
if [ -n "$SSHPASS" ] && command -v sshpass >/dev/null 2>&1; then
  SSH="sshpass -e ssh -o StrictHostKeyChecking=no"
fi

rsync -az --delete \
  --exclude '.git' \
  --exclude 'node_modules' \
  --exclude 'dist' \
  --exclude '.vite' \
  --exclude 'design_ref.png' \
  --exclude '.env' \
  --exclude 'backups' \
  --exclude 'uploads' \
  --exclude 'docker-compose.prod.yml' \
  --exclude '*.log' \
  -e "$SSH" ./ "$HOST:$TARGET/"

$SSH "$HOST" "cd $TARGET && docker compose -f docker-compose.prod.yml build backend frontend admin && docker compose -f docker-compose.prod.yml up -d && sleep 15 && docker compose -f docker-compose.prod.yml ps --format 'table {{.Name}}\t{{.Status}}'"

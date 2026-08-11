package migrations

import "embed"

// FS содержит SQL-миграции, вшитые в бинарь: на Railway нет ни entrypoint.sh,
// ни бинаря migrate, поэтому схему накатывает сам сервер при старте.
//
//go:embed *.sql
var FS embed.FS

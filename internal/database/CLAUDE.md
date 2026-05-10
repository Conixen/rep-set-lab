# internal/database — CLAUDE.md

## One file per entity

`users.go`, `workouts.go`, `exercises.go` — keep them separate.
Do not put queries for multiple entities in one file.

## Raw SQL only

No ORM. Use `sqlx.GetContext`, `sqlx.SelectContext`, `sqlx.QueryRowxContext`.
Struct field mapping is via `db:""` tags.

## Placeholder syntax

PostgreSQL uses `$1, $2, ...` — not `?`. Do not mix them.

## Migrations

`db.go` contains `Migrate()` which reads and executes all `migrations/*.sql` in
lexicographic order on every boot. Files must be idempotent (PostgreSQL IF NOT EXISTS guards).
New migration naming: `YYYYMMDDHHMMSS-description.sql`

## NULL handling

Use `sql.NullString` / `sql.NullTime` for nullable columns.
Do not use `*string` — it breaks sqlx struct scanning with PostgreSQL.

# rep-set-lab — CLAUDE.md

AI-powered gym training app backend. Go REST API with JWT auth, PostgreSQL, WebSockets,
and a three-way AI comparison layer (Claude / GPT-4o / Gemini).

## Tech stack

| Layer      | Choice                                  |
|------------|-----------------------------------------|
| HTTP       | Gin                                     |
| DB         | PostgreSQL + sqlx (raw SQL, no ORM)     |
| Auth       | JWT (golang-jwt/jwt v5) + bcrypt        |
| AI         | Claude, OpenAI, Gemini via official SDKs|
| WebSocket  | gorilla/websocket                       |
| Config     | godotenv + struct-based (no flag.*globals)|

## Project layout

```
cmd/api/          — entry point, wires everything, starts Gin
internal/
  auth/           — JWT middleware + register/login handlers
  ai/             — Provider interface, Claude/OpenAI/Gemini implementations, compare handler
  workout/        — generation + completion service + HTTP handlers
  user/           — stats handler
  exercise/       — exercise library handler + seed data
  xp/             — XP calculation and level threshold logic
  ws/             — WebSocket hub (real-time XP push)
  database/       — sqlx stores (one file per entity), migration runner
  config/         — struct-based config loaded from env
  validate/       — field-level validation helpers (RequiredString, PositiveInt, etc.)
  cache/          — generic getWithCache[T] helper (wired up when Redis is added)
  mock/           — handwritten mocks for ai.Provider and workout.Storage
migrations/       — idempotent PostgreSQL SQL files (IF NOT EXISTS guards)
logs/             — daily dev reports (YYYY-MM-DD.txt)
```

## Validation rule

Every handler that accepts user input must validate with the helpers in `internal/validate/`.
Error messages must name the exact field and state the valid values:

```go
// Good
validate.PositiveInt("duration_minutes", req.DurationMinutes)
// → "field 'duration_minutes' must be > 0, got -1"

// Bad
return errors.New("invalid input")
```

## Adding a new AI provider

1. Implement `ai.Provider` in `internal/ai/<name>.go`
2. Add a constructor `NewX(apiKey string) *XProvider`
3. Wire it in `cmd/api/main.go` under its env key check
4. Add the key to `.env.example`
5. The compare endpoint picks it up automatically

## XP system

- XP earned = `duration_minutes × 10` per completed workout
- Levels defined in `internal/xp/service.go` as `levelThresholds []int64`
- Extend the slice to add more levels — no other changes needed
- Future multipliers (streaks, difficulty) slot in via `Earned()` function

## Migrations

Files in `migrations/` run on every boot via `database.Migrate()`.
All files use PostgreSQL-idiomatic idempotency (`DO $$ IF NOT EXISTS ... END $$;`).
Name new files: `YYYYMMDDHHMMSS-description.sql`

## WebSocket

Connect: `GET /ws?token=<jwt>`
Server pushes JSON events on workout completion:
```json
{"type": "xp_update", "data": {"xp_earned": 600, "total_xp": 1100, "level": 2, "leveled_up": true}}
```

## Running locally

```bash
cp .env.example .env
# fill in DATABASE_URL and JWT_SECRET
# add at least one AI key
make run
```

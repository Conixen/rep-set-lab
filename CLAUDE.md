# rep-set-lab — CLAUDE.md

AI-powered gym training app — Go REST API backend with a Vue 3 frontend.

## Tech stack

| Layer      | Choice                                   |
|------------|------------------------------------------|
| HTTP       | Gin                                      |
| DB         | PostgreSQL + sqlx (raw SQL, no ORM)      |
| Auth       | JWT (golang-jwt/jwt v5) + bcrypt         |
| AI         | Claude, OpenAI, Gemini via official SDKs |
| WebSocket  | gorilla/websocket                        |
| Config     | godotenv + struct-based (no flag globals)|
| Frontend   | Vue 3 + Vite + TypeScript                |
| Styling    | Tailwind CSS v4                          |
| State      | Pinia                                    |
| Routing    | vue-router v4                            |

## Project layout

```
rep-set-lab/
├── backend/
│   ├── cmd/api/          — entry point, wires everything, starts Gin
│   ├── internal/
│   │   ├── auth/         — JWT middleware + register/login handlers
│   │   ├── ai/           — Provider interface, Claude/OpenAI/Gemini, compare handler
│   │   ├── workout/      — generation + completion service + HTTP handlers
│   │   ├── user/         — stats handler
│   │   ├── exercise/     — exercise library handler + seed data
│   │   ├── xp/           — XP calculation and level threshold logic
│   │   ├── ws/           — WebSocket hub (real-time XP push)
│   │   ├── database/     — sqlx stores (one file per entity), migration runner
│   │   ├── config/       — struct-based config loaded from env
│   │   ├── validate/     — field-level validation helpers
│   │   ├── cache/        — generic getWithCache[T] helper
│   │   └── mock/         — handwritten mocks for ai.Provider and workout.Storage
│   ├── migrations/       — idempotent PostgreSQL SQL files (IF NOT EXISTS guards)
│   ├── go.mod
│   ├── Makefile
│   └── .env.example
├── client/
│   ├── src/
│   │   ├── api/
│   │   │   └── client.ts — base fetch wrapper (auto-attaches JWT header)
│   │   ├── stores/
│   │   │   └── auth.ts   — Pinia auth store (login/register/logout)
│   │   ├── views/        — one file per route (Home, Library, Ai, Profile, Login)
│   │   ├── components/
│   │   │   └── icons/    — SVG icon components
│   │   ├── router/       — vue-router with auth guard
│   │   ├── App.vue       — shell + bottom navigation
│   │   └── main.ts
│   └── vite.config.ts    — proxies /api → localhost:8080 in dev
├── logs/                 — daily dev reports (YYYY-MM-DD.txt)
├── Makefile              — delegates to backend/ and client/
└── CLAUDE.md
```

## Frontend routes

| Path       | View           | Auth required |
|------------|----------------|---------------|
| `/login`   | LoginView      | No            |
| `/`        | HomeView       | Yes           |
| `/library` | LibraryView    | Yes           |
| `/ai`      | AiView         | Yes           |
| `/profile` | ProfileView    | Yes           |

## API base URL

All backend routes are prefixed `/api/v1`. In dev the Vite proxy forwards
`/api` → `http://localhost:8080`, so no CORS issues during development.

## Validation rule

Every handler that accepts user input must validate with the helpers in `backend/internal/validate/`.
Error messages must name the exact field and state the valid values:

```go
// Good
validate.PositiveInt("duration_minutes", req.DurationMinutes)
// → "field 'duration_minutes' must be > 0, got -1"

// Bad
return errors.New("invalid input")
```

## Adding a new AI provider

1. Implement `ai.Provider` in `backend/internal/ai/<name>.go`
2. Add a constructor `NewX(apiKey string) *XProvider`
3. Wire it in `backend/cmd/api/main.go` under its env key check
4. Add the key to `backend/.env.example`
5. The compare endpoint picks it up automatically

## XP system

- XP earned = `duration_minutes × 10` per completed workout
- Levels defined in `backend/internal/xp/service.go` as `levelThresholds []int64`
- Extend the slice to add more levels — no other changes needed

## Migrations

Files in `backend/migrations/` run on every boot via `database.Migrate()`.
All files use PostgreSQL-idiomatic idempotency (`DO $$ IF NOT EXISTS ... END $$;`).
Name new files: `YYYYMMDDHHMMSS-description.sql`

## WebSocket

Connect: `GET /ws?token=<jwt>`
Server pushes JSON on workout completion:
```json
{"type": "xp_update", "data": {"xp_earned": 600, "total_xp": 1100, "level": 2, "leveled_up": true}}
```

## Running locally

```bash
# Backend
cd backend
cp .env.example .env
# fill in DATABASE_URL, JWT_SECRET, and at least one AI key
make run          # → http://localhost:8080

# Frontend (separate terminal)
cd client
npm install
npm run dev       # → http://localhost:5173

# Or from the root (runs both in parallel)
make dev
```

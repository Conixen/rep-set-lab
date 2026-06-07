# Rep Set Lab

AI-powered gym training app. Generate personalized workouts, compare AI providers, and track your progress.

## Stack

- **Backend** — Go, Gin, PostgreSQL, sqlx
- **Frontend** — Vue 3, Vite, TypeScript, Tailwind CSS
- **AI** — Claude, OpenAI, Gemini, Groq

## Running locally

**Requirements:** Go 1.22+, Node 18+, PostgreSQL

```bash
# Clone and install
git clone https://github.com/Conixen/rep-set-lab
cd rep-set-lab

# Backend
cd backend
cp .env.example .env
# Fill in the required vars (see below), then:
make run

# Frontend (separate terminal)
cd client
npm install
npm run dev
```

App runs at `http://localhost:5173`, API at `http://localhost:8080`.

## Environment variables

### Required

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Long random string — use `openssl rand -hex 32` |

### AI providers (at least one required)

| Variable | Provider | Notes |
|---|---|---|
| `GEMINI_API_KEY` | Google Gemini | Free tier available — good starting point |
| `ANTHROPIC_API_KEY` | Claude | |
| `OPENAI_API_KEY` | OpenAI | |

Providers with an empty key are automatically excluded from workout generation and the compare view.

### Optional

| Variable | Description |
|---|---|
| `GROQ_API_KEY` | Groq | Used as the third-party grader in the compare feature |
| `EXERCISEDB_API_KEY` | RapidAPI key for ExerciseDB. Required for GIF previews in the exercise library. Without it the library still works but shows no images. Once set, trigger a sync from the admin panel (Admin → Exercises → Sync). Get a key at [rapidapi.com/justin-thewebdev/api/exercisedb](https://rapidapi.com/justin-thewebdev/api/exercisedb). |
| `BOOTSTRAP_ADMIN_EMAIL` | Email of the account to auto-promote to admin on every boot |
| `BOOTSTRAP_ADMIN_PASSWORD` | Password for the bootstrap admin account |
| `ALLOWED_ORIGINS` | Comma-separated allowed CORS origins. Defaults to `http://localhost:5173` in dev. Set to your frontend URL in production. |

## Project structure

```
rep-set-lab/
├── backend/
│   ├── cmd/api/          — entry point, wires everything, starts Gin
│   ├── internal/
│   │   ├── auth/         — JWT middleware, register/login handlers
│   │   ├── ai/           — Provider interface, Claude/OpenAI/Gemini/Groq
│   │   ├── workout/      — generation, completion, HTTP handlers
│   │   ├── exercise/     — library, seed data, GIF proxy, ExerciseDB sync
│   │   ├── admin/        — admin-only handlers
│   │   ├── xp/           — XP calculation and level logic
│   │   ├── ws/           — WebSocket hub for real-time XP updates
│   │   └── database/     — sqlx stores, migration runner
│   └── migrations/       — idempotent PostgreSQL SQL files
└── client/
    └── src/
        ├── api/          — base fetch wrapper with JWT handling
        ├── stores/       — Pinia stores (auth)
        ├── views/        — one file per route
        └── router/       — vue-router with auth guard
```

## Features

- Workout generation via multiple AI providers
- Side-by-side AI comparison with behavioral metrics and Groq as third party grader
- Exercise library with GIF previews
- XP and leveling system -  earn XP on completed workouts to track long-term progress.
- Real-time XP updates via WebSocket
- Admin panel for user management and data export

## Deployment

- **Backend + DB** — Railway
- **Frontend** — Vercel

Set `VITE_API_BASE_URL` in Vercel to your Railway backend URL.

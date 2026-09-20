# AGENTS.md — 3xN

Compact map for AI agents. Read this first; dig into code only for the task at hand.
Do **not** load all of `VISION.md` unless the task is product/vision related.

## Product

**3xN** — mobile training diary (PWA).

Preferred gym flow: pick a **cycle** → current **step** (session template) → log sets → finish → advance cycle cursor.

| Layer | Meaning |
|-------|---------|
| Exercise | Catalog item; kind `reps` \| `hold` + optional target / tabata protocol |
| Workout plan | Saved session template (ordered exercises) |
| Cycle | Named ordered steps of plans + cursor; home shows current step |
| Workout session | One performed visit (facts: sets, duration) |
| Interval protocol | Reusable tabata-like work/rest attached per exercise |

Product detail / backlog: [`VISION.md`](VISION.md).

## Stack

| Path | Role |
|------|------|
| [`diary-api/`](diary-api/) | Go + SQLite HTTP API (`:8080`) |
| [`diary-web/`](diary-web/) | Preact PWA + Vite (`:5173`), IndexedDB offline sync |

Email + password auth (`Bearer` session). Same-origin API via Vite proxy (`/v1`, `/healthz` → `127.0.0.1:8080`). Prefer empty `VITE_API_URL`.
All catalog/session rows are scoped by `user_id`. Offline IndexedDB is per user (`diary-offline-<userId>`).

## Layout (where to look)

### API (`diary-api/`)

- `cmd/server` — entry
- `internal/handlers` — HTTP (`router.go` = route list)
- `internal/repository` — SQLite
- `internal/models` — domain structs
- `migrations/` — SQL schema
- `data/diary.db` — local DB (do not commit secrets; backups under `data/backups/`)

### Web (`diary-web/`)

- `src/pages/` — screens (Router in `app.tsx`)
- `src/components/` — UI pieces (`BottomTabBar`, timers, banners)
- `src/api/client.ts` — fetch wrapper (+ ngrok skip header, Bearer token)
- `src/auth/session.ts` — localStorage token + user
- `src/sync/` — offline: IndexedDB, local workout, SyncWorker, catalog cache
- `src/types.ts` — shared TS types
- `public/` — manifest, icons, `tunnel-unlock.html`
- Dev footer: `DevLinksBar` (sync debug / tunnel / force update)

## Offline / sync (critical)

- Session id is always **client UUID** (`src/utils/id.ts`).
- Local-first workout → IndexedDB `workout_sessions` + ordered `sync_ops`.
- SyncWorker replays `start` → `saveExercise` → `finish` when online.
- `POST /v1/workout-sessions/start` requires `id`, idempotent.
- Catalog: **cache-first** (`sync/catalog.ts`), refresh in background.
- PWA: `vite-plugin-pwa`; phone updates need rebuild + SW refresh.

## Run locally

```bash
# API
cd diary-api && DIARY_BOOTSTRAP_PASSWORD='your-password' go run ./cmd/server

# Web (day-to-day HMR)
cd diary-web && npm run dev

# Web for iPhone / offline PWA (rebuilds each run)
cd diary-web && npm run start:pwa
# then: ngrok http 5173  → open https://….ngrok… on phone
```

Tests: `diary-web` → `npm test` (Vitest); API → Go tests under `internal/repository`.

### iPhone / ngrok notes

- Free ngrok may show **Visit Site**; unlock via footer «Подтвердить туннель» → `/tunnel-unlock.html` (PWA cookie jar).
- Stale UI: footer «Обновить приложение» or clear Safari site data for the ngrok host.
- Prefer `start:pwa` over `dev` for Home Screen icon testing.

## Agent conventions

- Communicate with the user in **Russian**; code comments / commits in **English**.
- Follow existing patterns (handlers ↔ repository; pages ↔ `api` + `sync`). Strict TypeScript; avoid `any`.
- Auth: `POST /v1/auth/register|login` public; everything else under `/v1` needs `Authorization: Bearer`.
- Do not invent magic-link/OAuth unless asked. Do not put passwords in git.
- Explain plan and wait for approval before large implementation (project preference).
- After API contract changes, provide a short Frontend Integration Card (method, path, types, sample `curl`).
- Do **not** commit unless the user asks. Do not invent calendars or math-engine HTTP.
- Prefer small diffs; no drive-by refactors or unsolicited markdown dumps.

## Docs

| File | Use when |
|------|----------|
| [`README.md`](README.md) | Quick start |
| [`diary-api/README.md`](diary-api/README.md) | API endpoints / curl |
| [`diary-web/README.md`](diary-web/README.md) | PWA / ngrok |
| [`VISION.md`](VISION.md) | Product decisions & backlog |

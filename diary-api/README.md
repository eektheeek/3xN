# Diary API

Go + SQLite HTTP service for the training diary (v1).

## Run

From this directory:

```bash
go run ./cmd/server
```

Defaults:

| Env | Default | Meaning |
|-----|---------|---------|
| `DIARY_DB` | `data/diary.db` | SQLite file path |
| `DIARY_MIGRATIONS` | `migrations` | SQL migrations directory |
| `DIARY_ADDR` | `:8080` | Listen address |

## Health check

```bash
curl -s http://localhost:8080/healthz
```

Expected: `{"ok":true}`

## Exercises

Kinds: `reps` (default) or `hold`. Target fields: `sets`, `reps`, `holdSec`, `weightKg`, `assistKg` (`assist` only meaningful for `reps`).

```bash
# Create exercise (reps)
curl -s -X POST http://localhost:8080/v1/exercises \
  -H 'Content-Type: application/json' \
  -d '{"name":"Wide grip pull-up","muscleGroup":"back","kind":"reps","supportsAssist":true}'

# Create hold exercise
curl -s -X POST http://localhost:8080/v1/exercises \
  -H 'Content-Type: application/json' \
  -d '{"name":"Single-leg bridge","muscleGroup":"core","kind":"hold"}'

# Set target (use id from create)
curl -s -X PUT http://localhost:8080/v1/exercises/EXERCISE_ID/target \
  -H 'Content-Type: application/json' \
  -d '{"sets":3,"reps":12,"holdSec":0,"weightKg":0,"assistKg":25}'

# Hold target example
curl -s -X PUT http://localhost:8080/v1/exercises/EXERCISE_ID/target \
  -H 'Content-Type: application/json' \
  -d '{"sets":3,"reps":0,"holdSec":60,"weightKg":0,"assistKg":0}'
```

## Workout sessions

Preferred gym flow is **incremental**:

1. `POST /v1/workout-sessions/start` with a **required client UUID** `id` (idempotent if replayed)
2. `PUT /v1/workout-sessions/{id}/exercises/{exerciseId}` to upsert sets
3. `POST /v1/workout-sessions/{id}/finish` with `durationSec` and optional `cycleId` (advances cycle when it matches the current step)

There is also a one-shot `POST /v1/workout-sessions` that creates a finished session with all exercises in one body (legacy / bulk).

```bash
# Start (client-generated UUID required)
curl -s -X POST http://localhost:8080/v1/workout-sessions/start \
  -H 'Content-Type: application/json' \
  -d '{"id":"'"$(uuidgen | tr '[:upper:]' '[:lower:]')"'","performedAt":"2026-07-22T10:00:00Z","workoutPlanId":"PLAN_A"}'

# Same id again → 200, no duplicate row
curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/v1/workout-sessions/start \
  -H 'Content-Type: application/json' \
  -d '{"id":"SESSION_UUID","performedAt":"2026-07-22T10:00:00Z","workoutPlanId":"PLAN_A"}'

# Save one exercise block
curl -s -X PUT http://localhost:8080/v1/workout-sessions/SESSION_UUID/exercises/EXERCISE_ID \
  -H 'Content-Type: application/json' \
  -d '{"position":1,"sets":[{"setNumber":1,"reps":12,"durationSec":0,"weightKg":0,"assistKg":25}]}'

# Finish (+ optional cycle advance)
curl -s -X POST http://localhost:8080/v1/workout-sessions/SESSION_UUID/finish \
  -H 'Content-Type: application/json' \
  -d '{"durationSec":3600,"cycleId":"CYCLE_ID"}'

# List / get
curl -s http://localhost:8080/v1/workout-sessions
curl -s http://localhost:8080/v1/workout-sessions/SESSION_UUID
```

Offline-capable web clients generate `id` locally, queue `start` / `saveExercise` / `finish`, then replay these endpoints when online.

## One-shot session (optional)

```bash
curl -s -X POST http://localhost:8080/v1/workout-sessions \
  -H 'Content-Type: application/json' \
  -d '{
    "performedAt":"2026-07-16T18:00:00Z",
    "isDeload":false,
    "exercises":[{
      "exerciseId":"EXERCISE_ID",
      "sets":[
        {"setNumber":1,"reps":12,"durationSec":0,"weightKg":0,"assistKg":25},
        {"setNumber":2,"reps":12,"durationSec":0,"weightKg":0,"assistKg":25},
        {"setNumber":3,"reps":12,"durationSec":0,"weightKg":0,"assistKg":25}
      ]
    }]
  }'
```

## Training cycles

Multiple named cycles (e.g. home / outdoor), each with its own ordered steps and cursor.

```bash
# Create cycles
curl -s -X POST http://localhost:8080/v1/cycles \
  -H 'Content-Type: application/json' \
  -d '{"name":"Дом"}'

curl -s -X POST http://localhost:8080/v1/cycles \
  -H 'Content-Type: application/json' \
  -d '{"name":"Улица"}'

# List / get
curl -s http://localhost:8080/v1/cycles
curl -s http://localhost:8080/v1/cycles/CYCLE_ID

# Set ordered steps
curl -s -X PUT http://localhost:8080/v1/cycles/CYCLE_ID/steps \
  -H 'Content-Type: application/json' \
  -d '{"workoutPlanIds":["PLAN_A","PLAN_B","PLAN_C"]}'

# Preferred: finish session and advance cycle in one request
curl -s -X POST http://localhost:8080/v1/workout-sessions/SESSION_ID/finish \
  -H 'Content-Type: application/json' \
  -d '{"durationSec":3600,"cycleId":"CYCLE_ID"}'

# Manual advance (recovery / debug)
curl -s -X POST http://localhost:8080/v1/cycles/CYCLE_ID/advance \
  -H 'Content-Type: application/json' \
  -d '{"sessionId":"SESSION_ID"}'
curl -s -X POST http://localhost:8080/v1/cycles/CYCLE_ID/restart
curl -s -X POST http://localhost:8080/v1/cycles/CYCLE_ID/repeat
curl -s -X PUT http://localhost:8080/v1/cycles/CYCLE_ID/on-home \
  -H 'Content-Type: application/json' \
  -d '{"onHome":true}'
```

## Tests

```bash
go test ./...
```

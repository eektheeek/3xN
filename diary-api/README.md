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

## End-to-end curl (pull-ups + target + log)

```bash
# 1) Create exercise
curl -s -X POST http://localhost:8080/v1/exercises \
  -H 'Content-Type: application/json' \
  -d '{"name":"Wide grip pull-up","muscleGroup":"back","supportsAssist":true}'

# 2) Set target (use id from step 1)
curl -s -X PUT http://localhost:8080/v1/exercises/EXERCISE_ID/target \
  -H 'Content-Type: application/json' \
  -d '{"sets":3,"reps":12,"weightKg":0,"assistKg":25}'

# 3) Save workout result
curl -s -X POST http://localhost:8080/v1/workout-sessions \
  -H 'Content-Type: application/json' \
  -d '{
    "performedAt":"2026-07-16T18:00:00Z",
    "isDeload":false,
    "exercises":[{
      "exerciseId":"EXERCISE_ID",
      "sets":[
        {"setNumber":1,"reps":12,"weightKg":0,"assistKg":25},
        {"setNumber":2,"reps":12,"weightKg":0,"assistKg":25},
        {"setNumber":3,"reps":12,"weightKg":0,"assistKg":25}
      ]
    }]
  }'

# 4) Read session (use id from step 3)
curl -s http://localhost:8080/v1/workout-sessions/SESSION_ID
```

## Tests

```bash
go test ./...
```

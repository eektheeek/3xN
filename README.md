# Dead Lift Project

Training diary product. Product vision: [VISION.md](VISION.md).

## Current focus

**Diary API (v1)** — Go + SQLite HTTP service:

- create an exercise
- set a target (sets / reps / load)
- save a workout session result

The previous `math-engine` package has been removed. Service: [`diary-api/`](diary-api/).

## Layout

| Path | Role |
|------|------|
| `VISION.md` | Product vision |
| `diary-api/` | HTTP diary API (Go + SQLite) |

```bash
cd diary-api && go run ./cmd/server
```

See [diary-api/README.md](diary-api/README.md).

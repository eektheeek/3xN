# Dead Lift Project

Training diary product. Product vision: [VISION.md](VISION.md).

Session templates (`workout_plans`) and named **training cycles** (ordered templates + current step, e.g. home vs outdoor) — see vision. Recovery windows come later.

## Stack

| Path | Role |
|------|------|
| [diary-api/](diary-api/) | Go + SQLite HTTP API |
| [diary-web/](diary-web/) | Preact PWA (mobile) |

## Quick start

```bash
# API
cd diary-api && go run ./cmd/server

# Web (another terminal)
cd diary-web && npm install && npm run dev
```

- API: http://localhost:8080  
- Web: http://localhost:5173  

See [diary-api/README.md](diary-api/README.md) and [diary-web/README.md](diary-web/README.md).

## iPhone

Use Safari → open `http://<mac-ip>:5173` → Share → **Add to Home Screen**.  
Set `VITE_API_URL=http://<mac-ip>:8080` in `diary-web/.env` so the phone can reach the API.

# Diary Web

Mobile PWA built with Preact. Talks to [diary-api](../diary-api).

## Setup

```bash
cd diary-web
npm install
```

By default the app calls `/v1/...` on the **same origin**. Vite proxies those to `http://127.0.0.1:8080` (see `vite.config.ts`). You usually do **not** need `VITE_API_URL` locally.

Optional `.env` only if the API is on another host (e.g. production):

```
VITE_API_URL=https://api.example.com
```

```bash
# terminal 1
cd diary-api && go run ./cmd/server

# terminal 2
cd diary-web && npm run dev
```

## Screens

| Route | Purpose |
|-------|---------|
| `/` | Saved workout plans |
| `/workouts/new` | Build a workout: name + exercises from catalog or new ones |
| `/workouts/:id` | Plan contents → **Start** |
| `/workouts/:id/start` | Log sets for all exercises and save |
| `/exercises/new` | Create an exercise (returns to plan builder) |
| `/exercises/:id` | Set target (sets / reps / weight / band assist) |

## iPhone (offline / PWA)

iOS only allows Service Workers on **HTTPS** (or localhost). Plain `http://<lan-ip>:5173` will load online but fail offline from the Home Screen icon.

Both `npm run dev` and `npm run preview` / `npm run start:pwa` use **port 5173** (`strictPort`).

### Recommended: ngrok + production build

```bash
# terminal 1 — API
cd diary-api && go run ./cmd/server

# terminal 2 — front with SW (rebuilds every time you run it)
cd diary-web && npm run start:pwa

# terminal 3 — HTTPS tunnel to the front only
ngrok http 5173
```

After each code change on Mac: **stop** `start:pwa` (Ctrl+C) → run `npm run start:pwa` again.
`npm run dev` does **not** reliably update the Home Screen PWA (old SW cache).

On the phone, if the UI looks stale: tap **«Обновить приложение»** (clears SW + caches), or open the ngrok URL in Safari and refresh twice.

1. On the phone open the **https://….ngrok…** URL (not the LAN IP).
2. Load the home screen once online (catalog → IndexedDB, SW installs).
3. Share → **Add to Home Screen** from that ngrok URL.
4. Airplane mode → reopen the icon → app shell should open; start a cached workout.

Remove any old Home Screen icon that pointed at `http://<ip>:5173`.

- Day-to-day coding with HMR: `npm run dev` (Safari tab is fine; SW still off in dev).
- Offline check: `npm run start:pwa` + ngrok as above.

If the ngrok URL changes, add a new Home Screen icon.

Without a prior online visit, offline will not work (SW not installed yet / empty catalog).

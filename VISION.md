# Product Vision (living document)

This file captures the global product idea. Extend it when new decisions or ideas appear.

## Core idea

The user builds reusable **session templates** (ordered exercises + optional interval protocols) and assembles them into one or more **training cycles** — sequential tracks of workouts (Step 1…N), e.g. “Дом” and “Улица”. At the gym the preferred path is: pick a cycle → see the **current step** → run it. Skipping ahead is not part of the main flow; if the user rests for a week, the same current step simply waits.

The library of saved templates remains available: any template can still be opened outside a cycle cursor when needed.

## Gym-day flow

1. Open app before or at the gym.
2. Prefer **current step of a chosen cycle** (progress “Step X of N” + Start). Optionally open any saved session template from the library.
3. Each exercise already has what it needs:
   - **reps** kind: sets × reps + optional weight / assist,
   - **hold** kind: sets × target hold seconds (static / mobility); log actual seconds per set,
   - optional interval protocol (tabata) on either kind.
4. Run timers, log sets (reps or seconds), finish session.
5. System records facts; when the finished session was the **current step of that cycle**, advance the cursor. Analytics / next recommendations come later.

## Planning vs execution

| Layer | Purpose |
|-------|---------|
| **Exercise** | Movement with kind `reps` \| `hold` + optional `IntervalProtocol` (tabata still attachable to either) |
| **Session template** (`workout_plan`) | Saved ordered stack of exercises for one gym visit (reusable) |
| **Cycle** | Named ordered list of session templates + **cursor** (current step). Anti-calendar: no Mon/Wed grid. Multiple cycles OK (home / outdoor) |
| **Session** | One gym visit fact: logged sets, duration, timestamps |
| **Math engine** | Pure calculation: metrics, fatigue, weight recommendations — no UI/DB (future) |

## Interval protocols (tabata-like)

- Protocols are **per exercise**, not per whole session.
- User can save and reuse protocols (`protocolId` + `displayName`) — no re-entering work/rest every time.
- Number of work intervals = number of sets logged; rest comes from protocol settings.
- Users who do not use intervals: `intervalProtocol` is omitted — not blocked.

## Personal constraints (v1 persona, extensible later)

- ~3 sessions per week, recovery-first.
- Progression: `% of e1RM` with adaptive load reduction when under-recovered (future).
- Deload / recovery **windows** and light/rehab **branches** (git-like merge) — future; not required for program-stack assembly.

## What we are NOT building in math-engine

- HTTP API, database, auth, UI screens.
- Those layers consume `schema/core_input_v1` and `schema/core_output_v1` later.

## UI: workout builder (stack model)

Two stack levels. The session level already exists in product; **cycles** are the next build.

### Exercise library

- Dedicated **exercise catalog** (not only ad-hoc names per session).
- Exercises always at hand when building a workout: search, filter, favorites.
- Catalog is shared across planning and gym execution.
- **v1.1 kinds:** `reps` (strength / bodyweight with optional assist) and `hold` (timed holds; target and log in seconds; session stats = sum of hold seconds).

### Session stack (builder UX) — existing

- User adds exercises from the catalog into an ordered list for one training session.
- Order in the stack = execution order in the gym.
- Per stack item (exercise):
  - attach optional **saved tabata / interval protocol**,
  - or leave without timer (strength-style sets only).
- A completed session stack is a **saved workout template** (`workout_plan`).

### Training cycle (builder UX) — next

- User creates named cycles and orders saved session templates into Steps 1…N.
- Home can highlight the **current** step of a selected cycle; finishing it advances that cycle’s cursor.
- Diary calendar stays a **history** of performed sessions, not a week planner.

### Gym execution UI

- Open current (or chosen) template → run through exercises in order.
- Timers from attached protocols; log sets/weight as you go.
- Math engine consumes session facts after completion (later).

## Open backlog (ideas to refine)

- [ ] **Recovery windows** (green / yellow / red by days since last session)
- [ ] **Light / rehab branches** + merge back to main track
- [ ] Auto progression of targets after successful main-track steps
- [ ] Auto rest timer after checking off a strength set
- [ ] Offline-first session logging + sync
- [ ] “Suggested protocols” from history
- [x] Multiple active UI affordances for switching cycles on home
- [x] Workout builder UI: session stack from catalog + per-exercise tabata
- [x] Exercise catalog DB/API
- [x] Saved session templates (`workout_plans`)
- [ ] Cycle builder: assemble templates + cursor + home “current step”

## Changelog (vision)

| Date | Note |
|------|------|
| 2026-05-28 | Initial vision: plan ahead, execute in gym, saved interval protocols, math-engine separate |
| 2026-05-28 | UI builder: exercise catalog, session stack, attach tabata per exercise, save stacks for multi-week planning |
| 2026-07-20 | Added **training cycles** (multi, named) on top of session templates; preferred gym path = current step of a cycle (anti-calendar). Diary remains history. Recovery windows / branches stay in backlog |
| 2026-07-20 | Renamed program→cycle; removed singleton `main` id |
| 2026-07-21 | **v1.1:** exercise kinds `reps` \| `hold`. Hold = multi-set timed goals (e.g. 3×60s) with optional load weight; log actual seconds (+ kg) per set; stats = total hold time as ч/м/с. Assist stays for reps. Tabata attachable as before. No dedicated hold start/stop timer in v1.1 (manual seconds). |

# Product Vision (living document)

This file captures the global product idea. Extend it when new decisions or ideas appear.

## Core idea

The user plans training in advance (week or month) from a library of exercises and saved interval protocols. At the gym they do not think about structure — they open the app and either follow the planned session or pick which planned workout to run today.

## Gym-day flow

1. Open app before or at the gym.
2. See today’s planned workout (or choose one from the weekly/monthly plan).
3. Each exercise already has what it needs:
   - load / reps guidance (from math engine),
   - interval protocol where relevant (warm-up, static holds, main work, tabata-style blocks).
4. Run timers, log sets, finish session.
5. System records facts and updates analytics / next recommendations.

## Planning vs execution

| Layer | Purpose |
|-------|---------|
| **Plan** | Week/month program: which exercises, which days, which saved protocols |
| **Session** | One gym visit: ordered exercises with pre-attached protocols |
| **Exercise** | Movement + sets + optional `IntervalProtocol` (work/rest, `protocolId`, `displayName`) |
| **Math engine** | Pure calculation: metrics, fatigue, weight recommendations — no UI/DB |

## Interval protocols (tabata-like)

- Protocols are **per exercise**, not per whole session.
- User can save and reuse protocols (`protocolId` + `displayName`) — no re-entering work/rest every time.
- Number of work intervals = number of sets logged; rest comes from protocol settings.
- Users who do not use intervals: `intervalProtocol` is omitted — not blocked.

## Personal constraints (v1 persona, extensible later)

- ~3 sessions per week, recovery-first.
- Progression: `% of e1RM` with adaptive load reduction when under-recovered.
- Deload: `3 load weeks + 1 deload week` (configurable via policy, not hardcoded to one user in code).

## What we are NOT building in math-engine

- HTTP API, database, auth, UI screens.
- Those layers consume `schema/core_input_v1` and `schema/core_output_v1` later.

## UI: workout builder (stack model)

Future client UI is required to compose workouts without manual re-planning every day.

### Exercise library (separate database)

- Dedicated **exercise catalog** (not only ad-hoc names per session).
- Exercises always at hand when building a workout: search, filter, favorites.
- Catalog is shared across planning and gym execution (same IDs as in math-engine input).

### Session stack (builder UX)

- Planning screen behaves like a **stack**: user adds exercises from the catalog into an ordered list for one training session.
- Order in the stack = execution order in the gym.
- Per stack item (exercise):
  - attach optional **saved tabata / interval protocol** (warm-up, static, main work, finisher),
  - or leave without timer (strength-style sets only).
- Visual metaphor: “pile” exercises on top of each other → one ready session template.

### Saved stacks and multi-week planning

- A completed stack is a **saved workout template** (reusable session definition).
- User can assign saved stacks to calendar slots (week or month view).
- Goal: plan several weeks ahead once, then at the gym only pick “today’s stack” or override if needed.
- No need to re-enter work/rest or hunt exercises each visit.

### Gym execution UI (ties to stack)

- Open planned stack → run through exercises in order.
- Timers pre-filled from attached protocols; log sets/weight as you go.
- Math engine consumes session facts after completion.

## Open backlog (ideas to refine)

- [ ] Program templates (4–6 week blocks) vs ad-hoc weekly plans
- [ ] “Suggested protocols” from history (most used work/rest pairs)
- [ ] Offline-first session logging + sync
- [ ] Minimal UI: timer, set entry, one analytics screen
- [ ] **Workout builder UI**: stack of exercises from catalog + per-exercise tabata attach
- [ ] **Exercise catalog DB/API**: CRUD, tags, muscle groups, user favorites
- [ ] **Saved stack templates**: save, duplicate, assign to week/month calendar
- [ ] Calendar view: map stacks → training days for 1–4+ weeks

## Changelog (vision)

| Date | Note |
|------|------|
| 2026-05-28 | Initial vision: plan ahead, execute in gym, saved interval protocols, math-engine separate |
| 2026-05-28 | UI builder: exercise catalog, session stack, attach tabata per exercise, save stacks for multi-week planning |

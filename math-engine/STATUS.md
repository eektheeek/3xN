# Math Engine — Save State

**Updated:** 2026-05-28  
**Rule:** one step per session → implement → verify → stop → wait for explicit "go next".

Product vision (UI, stacks, catalog): [../VISION.md](../VISION.md)

---

## Done (this sprint slice)

| Area | Artifact |
|------|----------|
| Layout | `math-engine/` module (`go.mod`: `github.com/eektheeek/dead-lift-project/math-engine`) |
| Schemas | `schema/core_input_v1.json`, `core_output_v1.json`, `field_notes.md` |
| Entities | `workout_session.go`, `training_config.go`, `user_metrics.go` |
| Formulas | `epley.go`, `brzycki.go`, `lombardi.go`, `ensemble.go` (median, skip invalid `0`) |

### Locked decisions

- Inside-Out: math only, no HTTP/DB/UI in this module.
- Intervals on **Exercise**; `len(sets)` = work intervals; no `rounds` / `intervalExecution`.
- `IntervalProtocol`: `protocolId`, `workSec`, `restSec`, `displayName`.
- Progression/deload: `strategy` field (`percent_e1rm`, `fixed_3_plus_1`).
- No `sessionDurationMin` in constraints; no Mayhew in v1.
- e1RM: **median** of Epley + Brzycki + Lombardi.

---

## Next steps (strict order)

Each item = **one** change set, then pause.

### Stage 1.1 (finish domain input)

1. [x] `internal/entities/user_metrics.go` — `UserMetrics` (age, body weight, recovery sensitivity, `fatigueThreshold`, `readinessThreshold`).
2. [x] Align `core_input_v1.json` vs entities (see sync report below).
3. [x] `internal/validate/` — `go-playground/validator` + `contracts/input.go` (tags, no hand-written if chain).

### Stage 1.2 (metrics, still no pipeline)

4. [x] `internal/metrics/volume.go` — `VolumeLoad = Σ(weight × reps)`.
5. [x] `internal/formulas/ensemble_test.go` — golden: 50×8 → ~62.1 median.

### Stage 1.3–1.4 (recommendation)

6. [x] `internal/recommendation/` — `targetWeight = e1RM × targetPercent × modifiers`; actions `increase|hold|decrease`.
6b. [x] Modifier thresholds in `userMetrics` (`fatigueThreshold`, `readinessThreshold`; `0` → defaults 0.96 / 0.75).
7. [x] `internal/recommendation/deload.go` — `IsDeloadWeek(policy, weekIndex)` (1-based cycle; presets 3+1, 2+1, custom).

### Stage 1.5 (CLI)

8. [x] `cmd/core-cli/` — `validate-input`, `run-analysis`; `internal/analysis`, `internal/normalize`, `contracts/output.go`.

### Later (not now)

- `internal/timer/` — timeline from `IntervalProtocol` + sets count.
- Foster metrics (monotony, strain, ACWR).
- Stage 2+ backend / UI per [VISION.md](../VISION.md).

---

## Resume command

```bash
cd math-engine && go build ./...
```

Next action: **Stage 1 wrap** — Foster/timer, or normalize hardening; then Stage 2 backend per VISION.

### Sync report (step 2, 2026-05-28)

| JSON path | Go entity | Status |
|-----------|-----------|--------|
| `rawTrainingLog.*` | `WorkoutSession` | aligned (session only) |
| `userMetrics` (top-level) | `UserMetrics` | aligned |
| `exercises[]` | `Exercise`, `SetEntry` | aligned |
| `intervalProtocol` | `IntervalProtocol` | aligned (optional per exercise) |
| `userMetrics` | `UserMetrics` | aligned |
| `trainingConstraints` | `TrainingConstraints` | aligned |
| `progressionPolicy` | `ProgressionPolicy` | aligned (`strategy`) |
| `deloadPolicy` | `DeloadPolicy` | aligned (`strategy`) |
| Top-level `CoreInput` | `contracts.CoreInput` | aligned (`validate` tags) |

---

## Changelog (save state)

| Date | Note |
|------|------|
| 2026-05-28 | Initial save state after schemas, entities, e1RM formulas + ensemble |

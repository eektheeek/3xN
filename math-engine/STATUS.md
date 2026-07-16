# Math Engine — Save State

**Updated:** 2026-06-03  
**Rule:** one step per session → implement → verify → stop → wait for explicit "go next".

Product vision (UI, stacks, catalog): [../VISION.md](../VISION.md)

---

## Focus: Smart Trainer only

- **In:** `linear_step` progression, `exerciseOutcome`, volume, deload week, session modifiers.
- **Out:** per-exercise `smartTrainer` (working), stack order, `message` for UI reminders.
- **Removed (v1):** `internal/formulas/`, `internal/recommendation/`, e1RM in output, `percent_e1rm` strategy.
- **Later:** e1RM analytics can return as a separate optional module.

---

## Done

| Area | Artifact |
|------|----------|
| Schemas | `core_input_v1.json`, `core_output_v1.json`, `smart_trainer_rules.md` |
| CLI | `validate-input`, `run-analysis` |
| `internal/smarttrainer/` | `Run`, volume, `ForecastSession`, `SuggestNext`, `IsDeloadWeek`, tests |
| Contracts | `exerciseType`, `exerciseOutcome`, `SmartTrainerOut` |

### Smart trainer track

- [x] ST.1 — `exerciseOutcome` / `exerciseType`
- [x] ST.2 — `internal/smarttrainer/`
- [x] ST.3–4 — `smarttrainer.Run` + output `smartTrainer`; e1RM removed; `internal/analysis/` removed

---

## Next (Later)

- `internal/timer/` — interval timeline
- Foster / trends
- Backend + UI journal (store last `suggestedNextWeightKg` for pre-workout screen)
- Optional e1RM analytics package (not tied to load suggestion)

---

## Resume

```bash
cd math-engine && go test ./...
go run ./cmd/core-cli run-analysis -f schema/core_input_v1.json
```

Example working line: `suggestedNextWeightKg: 102.5` after 3×100×5 with `planCompleted` + `readyToProgress`.

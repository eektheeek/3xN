# Field Notes (core v1)

## core_input_v1
- `version` (required): Contract version, always `core_input_v1`.
- `rawTrainingLog` (required): Session payload → maps to `WorkoutSession`.
  - `sessionId`, `userId`, `startedAt`, `completedAt`, `sessionDurationSec`, `exercises`
  - `exercises[]`: `exerciseId`, `name`, `muscleGroup`, `exerciseType` (`warmup` | `working`), `sets[]`, optional `intervalProtocol`
  - `intervalProtocol`: `protocolId`, `workSec`, `restSec`, `displayName` (no `rounds`; interval count = `len(sets)`)
- `userMetrics` (required): Profile snapshot → `UserMetrics` (`age`, `bodyWeightKg`, `recoverySensitivity`, `fatigueThreshold`, `readinessThreshold`; use `0` for product defaults).
- `trainingConstraints` (required): `sessionsPerWeek` only.
- `progressionPolicy` (required): `strategy` (`linear_step` | `rir_target`), `increaseStepKg`, `decreaseStepKg`.
- `deloadPolicy` (required): user preset `strategy` (`fixed_3_plus_1` | `fixed_2_plus_1` | `custom`), weeks and drop % in allowed ranges.
- `trainingWeekIndex` (required): 1-based week in deload cycle (caller computes from plan/calendar).
- `sessionModifiers` (required): `fatigueModifier`, `readinessModifier` in `(0, 1]` for this session.
- `exerciseOutcome` (required for `working` only): `planCompleted`, `readyToProgress` — how the exercise went (see `schema/smart_trainer_rules.md`).

## core_output_v1
- `version` (required): `analysis_result_v1`.
- `sessionId`, `sessionDate`, `userId` (required): Correlation keys.
- `computedMetrics.volumeLoad` (required): Working-block tonnage only (`kg_reps`).
- `exerciseMetrics[]` (required): Stack order; warmup = ids only; working = `volumeLoad` + `smartTrainer`.
- `smartTrainer` (working only): `lastWorkingWeightKg`, `suggestedNextWeightKg`, `action`, `message`, `reasonCodes`.
- `sessionContext` (required): `fatigueModifier`, `readinessModifier`, `deloadActive`, `trainingWeekIndex`.
- `createdAt` (required): UTC RFC3339.

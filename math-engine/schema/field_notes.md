# Field Notes (core v1)

## core_input_v1
- `version` (required): Contract version, always `core_input_v1`.
- `rawTrainingLog` (required): Session payload → maps to `WorkoutSession`.
  - `sessionId`, `userId`, `startedAt`, `completedAt`, `sessionDurationSec`, `exercises`
  - `exercises[]`: `exerciseId`, `name`, `muscleGroup`, `blockType` (`warmup` | `working`), `sets[]`, optional `intervalProtocol`. Warmup blocks appear in output stack only (no e1rm/volume/recommendation).
  - `intervalProtocol`: `protocolId`, `workSec`, `restSec`, `displayName` (no `rounds`; interval count = `len(sets)`)
- `userMetrics` (required): Profile snapshot → `UserMetrics` (`age`, `bodyWeightKg`, `recoverySensitivity`, `fatigueThreshold`, `readinessThreshold`; use `0` for product defaults).
- `trainingConstraints` (required): `sessionsPerWeek` only.
- `progressionPolicy` (required): user preset `strategy` (`percent_e1rm` | `rir_target`), plus numeric params.
- `deloadPolicy` (required): user preset `strategy` (`fixed_3_plus_1` | `fixed_2_plus_1` | `custom`), weeks and drop % in allowed ranges.
- `trainingWeekIndex` (required): 1-based week in deload cycle (caller computes from plan/calendar).
- `sessionModifiers` (required): `fatigueModifier`, `readinessModifier` in `(0, 1]` for this session (UI check-in or auto metrics later).

## core_output_v1
- `version` (required): Output contract version, `analysis_result_v1`.
- `sessionId` (required): Session identifier for correlation.
- `sessionDate` (required): Date key used for trend series.
- `userId` (required): User identifier.
- `computedMetrics` (required): Session-level metrics with confidence/applicability.
- `exerciseMetrics` (required): Per-exercise metrics including e1RM breakdown.
- `trendPoints` (optional): Time-series points for charts and trend analysis.
- `recommendation` (required): Next-session action and target load range.
- `intervalSummary` (optional): Compliance summary for timer protocols.
- `intervalEvents` (optional): Detailed timeline events for work/rest phases.
- `warnings` (optional): Data-quality or applicability warnings.
- `createdAt` (required): Analysis creation timestamp (UTC, RFC3339).

-- Snapshot exercise target on session blocks so diary history stays stable.

ALTER TABLE workout_session_exercises ADD COLUMN target_sets INTEGER;
ALTER TABLE workout_session_exercises ADD COLUMN target_reps INTEGER;
ALTER TABLE workout_session_exercises ADD COLUMN target_hold_sec INTEGER;
ALTER TABLE workout_session_exercises ADD COLUMN target_weight_kg REAL;
ALTER TABLE workout_session_exercises ADD COLUMN target_assist_kg REAL;

-- One-time backfill from current catalog targets (only rows without a snapshot).
UPDATE workout_session_exercises
SET
  target_sets = (SELECT t.target_sets FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id),
  target_reps = (SELECT t.target_reps FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id),
  target_hold_sec = (SELECT t.hold_sec FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id),
  target_weight_kg = (SELECT t.weight_kg FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id),
  target_assist_kg = (SELECT t.assist_kg FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id)
WHERE target_sets IS NULL
  AND EXISTS (
    SELECT 1 FROM exercise_targets t WHERE t.exercise_id = workout_session_exercises.exercise_id
  );

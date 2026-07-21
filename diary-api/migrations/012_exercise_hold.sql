-- v1.1: exercise kinds reps|hold; hold targets and set durations in seconds.

ALTER TABLE exercises ADD COLUMN kind TEXT NOT NULL DEFAULT 'reps'
    CHECK (kind IN ('reps', 'hold'));

ALTER TABLE exercise_targets ADD COLUMN hold_sec INTEGER NOT NULL DEFAULT 0
    CHECK (hold_sec >= 0);

ALTER TABLE sets ADD COLUMN duration_sec INTEGER NOT NULL DEFAULT 0
    CHECK (duration_sec >= 0);

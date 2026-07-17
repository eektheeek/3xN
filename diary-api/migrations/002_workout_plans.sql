-- Workout plans (saved templates the user can start).

CREATE TABLE IF NOT EXISTS workout_plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workout_plan_exercises (
    id               TEXT PRIMARY KEY,
    workout_plan_id  TEXT NOT NULL REFERENCES workout_plans(id) ON DELETE CASCADE,
    exercise_id      TEXT NOT NULL REFERENCES exercises(id),
    position         INTEGER NOT NULL,
    UNIQUE (workout_plan_id, position)
);

CREATE INDEX IF NOT EXISTS idx_workout_plan_exercises_plan
    ON workout_plan_exercises(workout_plan_id);

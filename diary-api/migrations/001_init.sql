-- Initial diary schema (v1).

CREATE TABLE IF NOT EXISTS exercises (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    muscle_group    TEXT NOT NULL DEFAULT '',
    supports_assist INTEGER NOT NULL DEFAULT 0 CHECK (supports_assist IN (0, 1)),
    created_at      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS exercise_targets (
    exercise_id  TEXT PRIMARY KEY REFERENCES exercises(id) ON DELETE CASCADE,
    target_sets  INTEGER NOT NULL,
    target_reps  INTEGER NOT NULL,
    weight_kg    REAL NOT NULL DEFAULT 0,
    assist_kg    REAL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS workout_sessions (
    id            TEXT PRIMARY KEY,
    performed_at  TEXT NOT NULL,
    is_deload     INTEGER NOT NULL DEFAULT 0 CHECK (is_deload IN (0, 1)),
    created_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workout_session_exercises (
    id                  TEXT PRIMARY KEY,
    workout_session_id  TEXT NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id         TEXT NOT NULL REFERENCES exercises(id),
    position            INTEGER NOT NULL,
    UNIQUE (workout_session_id, position)
);

CREATE TABLE IF NOT EXISTS sets (
    id                           TEXT PRIMARY KEY,
    workout_session_exercise_id  TEXT NOT NULL REFERENCES workout_session_exercises(id) ON DELETE CASCADE,
    set_number                   INTEGER NOT NULL,
    reps                         INTEGER NOT NULL,
    weight_kg                    REAL NOT NULL DEFAULT 0,
    assist_kg                    REAL NOT NULL DEFAULT 0,
    UNIQUE (workout_session_exercise_id, set_number)
);

CREATE INDEX IF NOT EXISTS idx_workout_session_exercises_session
    ON workout_session_exercises(workout_session_id);

CREATE INDEX IF NOT EXISTS idx_workout_session_exercises_exercise
    ON workout_session_exercises(exercise_id);

CREATE INDEX IF NOT EXISTS idx_sets_workout_session_exercise
    ON sets(workout_session_exercise_id);

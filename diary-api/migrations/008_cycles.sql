-- Training cycles: ordered workout_plans + per-cycle cursor.
-- Multiple cycles allowed (e.g. "Дом", "Улица").

CREATE TABLE IF NOT EXISTS cycles (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    current_step  INTEGER NOT NULL DEFAULT 1 CHECK (current_step >= 1),
    created_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cycle_steps (
    id               TEXT PRIMARY KEY,
    cycle_id         TEXT NOT NULL REFERENCES cycles(id) ON DELETE CASCADE,
    position         INTEGER NOT NULL CHECK (position >= 1),
    workout_plan_id  TEXT NOT NULL REFERENCES workout_plans(id),
    UNIQUE (cycle_id, position)
);

CREATE INDEX IF NOT EXISTS idx_cycle_steps_cycle
    ON cycle_steps(cycle_id);

ALTER TABLE workout_sessions ADD COLUMN workout_plan_id TEXT REFERENCES workout_plans(id);

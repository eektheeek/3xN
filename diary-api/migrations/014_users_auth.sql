-- Multi-user auth: users, sessions, and owner scoping on root tables.

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Empty string until bootstrap seed assigns the owner user id.
ALTER TABLE exercises ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE interval_protocols ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE workout_plans ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE cycles ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
ALTER TABLE workout_sessions ADD COLUMN user_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_exercises_user ON exercises(user_id);
CREATE INDEX IF NOT EXISTS idx_interval_protocols_user ON interval_protocols(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_plans_user ON workout_plans(user_id);
CREATE INDEX IF NOT EXISTS idx_cycles_user ON cycles(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_sessions_user ON workout_sessions(user_id);

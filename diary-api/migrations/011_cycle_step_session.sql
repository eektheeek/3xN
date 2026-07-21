-- Link finished cycle steps to diary sessions.
ALTER TABLE cycle_steps ADD COLUMN completed_session_id TEXT REFERENCES workout_sessions(id);

CREATE TABLE IF NOT EXISTS interval_protocols (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    work_sec      INTEGER NOT NULL CHECK (work_sec >= 1),
    rest_sec      INTEGER NOT NULL CHECK (rest_sec >= 0),
    warmup_extra  INTEGER NOT NULL DEFAULT 0 CHECK (warmup_extra IN (0, 1)),
    created_at    TEXT NOT NULL
);

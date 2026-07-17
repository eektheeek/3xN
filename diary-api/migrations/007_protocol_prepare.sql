ALTER TABLE interval_protocols ADD COLUMN prepare_sec INTEGER NOT NULL DEFAULT 5 CHECK (prepare_sec >= 0);

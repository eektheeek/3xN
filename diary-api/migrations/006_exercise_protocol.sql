ALTER TABLE exercises ADD COLUMN protocol_id TEXT REFERENCES interval_protocols(id) ON DELETE SET NULL;

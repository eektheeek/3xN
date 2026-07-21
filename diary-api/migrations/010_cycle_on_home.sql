-- Cycles visible on the home screen.
ALTER TABLE cycles ADD COLUMN on_home INTEGER NOT NULL DEFAULT 0 CHECK (on_home IN (0, 1));

-- Already progressed cycles appear on home.
UPDATE cycles SET on_home = 1 WHERE current_step > 1;

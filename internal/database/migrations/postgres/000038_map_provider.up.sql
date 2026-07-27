ALTER TABLE events ADD COLUMN map_provider TEXT NOT NULL DEFAULT 'google';
ALTER TABLE event_series ADD COLUMN map_provider TEXT NOT NULL DEFAULT 'google';

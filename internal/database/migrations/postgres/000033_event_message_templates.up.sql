CREATE TABLE event_message_templates (
    id           TEXT PRIMARY KEY,
    event_id     TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    message_type TEXT NOT NULL,
    subject      TEXT NOT NULL,
    body         TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    UNIQUE(event_id, message_type)
);

CREATE INDEX idx_event_message_templates_event_id ON event_message_templates(event_id);

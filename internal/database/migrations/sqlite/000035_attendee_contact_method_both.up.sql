-- SQLite does not support ALTER TABLE to modify CHECK constraints, so we
-- recreate the attendees table with 'both' added to the contact_method
-- CHECK constraint.

CREATE TABLE attendees_new (
    id             TEXT PRIMARY KEY,
    event_id       TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    email          TEXT,
    phone          TEXT,
    rsvp_status    TEXT NOT NULL DEFAULT 'pending' CHECK(rsvp_status IN ('pending','attending','maybe','declined','waitlisted')),
    rsvp_token     TEXT NOT NULL UNIQUE,
    contact_method TEXT NOT NULL DEFAULT 'email' CHECK(contact_method IN ('email','sms','both')),
    dietary_notes  TEXT NOT NULL DEFAULT '',
    plus_ones      INTEGER NOT NULL DEFAULT 0,
    import_source  TEXT,
    created_at     TEXT NOT NULL,
    updated_at     TEXT NOT NULL
);

INSERT INTO attendees_new SELECT * FROM attendees;

DROP TABLE attendees;

ALTER TABLE attendees_new RENAME TO attendees;

-- Recreate all indexes from migrations 000006, 000014, and 000022.
CREATE INDEX idx_attendees_event_id ON attendees(event_id);
CREATE INDEX idx_attendees_rsvp_token ON attendees(rsvp_token);
CREATE INDEX idx_attendees_email ON attendees(email);
CREATE INDEX idx_attendees_rsvp_status ON attendees(rsvp_status);
CREATE UNIQUE INDEX idx_attendees_event_email ON attendees(event_id, email) WHERE email != '';
CREATE UNIQUE INDEX idx_attendees_event_phone ON attendees(event_id, phone) WHERE phone != '';

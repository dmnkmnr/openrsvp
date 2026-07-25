-- PostgreSQL supports altering CHECK constraints in place, so we redefine the
-- inline column constraint (auto-named attendees_contact_method_check) to add
-- 'both' rather than rebuilding the table.
ALTER TABLE attendees DROP CONSTRAINT attendees_contact_method_check;
ALTER TABLE attendees ADD CONSTRAINT attendees_contact_method_check
    CHECK (contact_method IN ('email','sms','both'));

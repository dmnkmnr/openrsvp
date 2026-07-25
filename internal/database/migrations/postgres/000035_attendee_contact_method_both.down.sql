-- Revert the attendees.contact_method CHECK constraint to remove 'both'. Move
-- any rows out of that value first, then narrow the constraint.
UPDATE attendees SET contact_method = 'email' WHERE contact_method = 'both';
ALTER TABLE attendees DROP CONSTRAINT attendees_contact_method_check;
ALTER TABLE attendees ADD CONSTRAINT attendees_contact_method_check
    CHECK (contact_method IN ('email','sms'));

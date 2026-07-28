package rsvp

// ResolveChannels decides which channel(s) a notification should go out on,
// given the attendee's contactMethod preference and which contact info is
// actually on file. Never sends both email and SMS for the same
// notification: a guest with both on file gets email only (more reliable,
// avoids duplicate notifications), and SMS is reserved for guests who
// explicitly chose "sms" as their preference, or who only have a phone
// number at all. Every branch falls back to whichever channel actually has
// data when the preferred one doesn't (e.g. organizer removed an email
// after the guest chose "email").
func ResolveChannels(contactMethod string, hasEmail, hasPhone bool) (wantEmail, wantSMS bool) {
	switch contactMethod {
	case "sms":
		// Explicit SMS preference is honored even if the guest also has an
		// email on file.
		if hasPhone {
			return false, true
		}
		return hasEmail, false
	default: // "both", "email", or unset/legacy value
		if hasEmail {
			return true, false
		}
		return false, hasPhone
	}
}

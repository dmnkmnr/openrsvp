package rsvp

// ResolveChannels decides which channels a notification should go out on,
// given the attendee's contactMethod preference and which contact info is
// actually on file. "both" sends to whichever of email/phone is present
// (usually both); "email"/"sms" falls back to whichever channel actually
// has data when the preferred one doesn't (e.g. organizer removed an email
// after the guest chose "email").
func ResolveChannels(contactMethod string, hasEmail, hasPhone bool) (wantEmail, wantSMS bool) {
	switch contactMethod {
	case "both":
		return hasEmail, hasPhone
	case "sms":
		if hasPhone {
			return false, true
		}
		return hasEmail, false
	default: // "email" or unset/legacy value
		if hasEmail {
			return true, false
		}
		return false, hasPhone
	}
}

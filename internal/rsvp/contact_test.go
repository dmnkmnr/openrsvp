package rsvp

import "testing"

func TestResolveChannels(t *testing.T) {
	cases := []struct {
		contactMethod      string
		hasEmail, hasPhone bool
		wantEmail, wantSMS bool
	}{
		// "both": email only when both are present (never duplicate to SMS
		// too); falls back to whichever is actually present otherwise.
		{"both", true, true, true, false},
		{"both", true, false, true, false},
		{"both", false, true, false, true},
		{"both", false, false, false, false},

		// "email": prefer email, fall back to sms if no email.
		{"email", true, true, true, false},
		{"email", true, false, true, false},
		{"email", false, true, false, true},
		{"email", false, false, false, false},

		// "sms": prefer sms, fall back to email if no phone.
		{"sms", true, true, false, true},
		{"sms", true, false, true, false},
		{"sms", false, true, false, true},
		{"sms", false, false, false, false},

		// unset/legacy value: same fallback as "email".
		{"", true, true, true, false},
		{"", false, true, false, true},
	}

	for _, c := range cases {
		gotEmail, gotSMS := ResolveChannels(c.contactMethod, c.hasEmail, c.hasPhone)
		if gotEmail != c.wantEmail || gotSMS != c.wantSMS {
			t.Errorf("ResolveChannels(%q, %v, %v) = (%v, %v), want (%v, %v)",
				c.contactMethod, c.hasEmail, c.hasPhone, gotEmail, gotSMS, c.wantEmail, c.wantSMS)
		}
	}
}

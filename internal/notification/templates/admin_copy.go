package templates

// This file holds the fixed (non organizer-editable) localized copy for the
// organizer/admin-facing system emails -- magic-link sign-in, "new RSVP"
// notification, co-host invitation, retention warning, and feedback
// confirmation. Unlike the guest-facing notification templates, organizers
// cannot customize the wording of these, only the language they're rendered
// in (their own account language, internal/auth.Organizer.Language).

// --- Magic link ---

type magicLinkCopy struct {
	Subject     string
	Heading     string
	Intro       string
	ButtonLabel string
	// ExpiryNote has one %d verb for the expiry-minutes value.
	ExpiryNote string
}

var magicLinkCopyByLang = map[string]magicLinkCopy{
	"en": {
		Subject:     "Sign in to OpenRSVP",
		Heading:     "Sign in to your account",
		Intro:       "Click the button below to sign in to OpenRSVP. No password needed.",
		ButtonLabel: "Sign In",
		ExpiryNote:  "This link expires in %d minutes. If you did not request this link, you can safely ignore this email.",
	},
	"de": {
		Subject:     "Anmeldung bei OpenRSVP",
		Heading:     "Bei deinem Konto anmelden",
		Intro:       "Klicke auf den Button unten, um dich bei OpenRSVP anzumelden. Kein Passwort nötig.",
		ButtonLabel: "Anmelden",
		ExpiryNote:  "Dieser Link läuft in %d Minuten ab. Falls du diesen Link nicht angefordert hast, kannst du diese E-Mail ignorieren.",
	},
}

func magicLinkCopyFor(lang string) magicLinkCopy {
	if c, ok := magicLinkCopyByLang[lang]; ok {
		return c
	}
	return magicLinkCopyByLang[defaultLanguage]
}

// --- Organizer RSVP notification ---

type organizerRSVPNotificationCopy struct {
	SubjectPrefix string // e.g. "New RSVP"
	Heading       string
	// IntroFormat has one %s verb for the (HTML-escaped, bold-wrapped) event title.
	IntroFormat           string
	LabelGuest            string
	LabelResponse         string
	LabelEmail            string
	LabelPhone            string
	LabelAdditionalGuests string
	ChildrenSuffix        string // e.g. "under 12"
	ChildrenPlainFormat   string // e.g. "%d child(ren) under 12"
	ButtonLabel           string
	PlainHeading          string
	PlainLabelEvent       string
	PlainLabelGuest       string
	PlainLabelResponse    string
	PlainLabelEmail       string
	PlainLabelPhone       string
	PlainLabelAdditional  string
	PlainFooterCTA        string // e.g. "View your event dashboard:"
}

var organizerRSVPNotificationCopyByLang = map[string]organizerRSVPNotificationCopy{
	"en": {
		SubjectPrefix:         "New RSVP",
		Heading:               "New RSVP Received",
		IntroFormat:           "Someone has responded to your event <strong>%s</strong>.",
		LabelGuest:            "Guest",
		LabelResponse:         "Response",
		LabelEmail:            "Email",
		LabelPhone:            "Phone",
		LabelAdditionalGuests: "Additional Guests",
		ChildrenSuffix:        "under 12",
		ChildrenPlainFormat:   "%d child(ren) under 12",
		ButtonLabel:           "View Event Dashboard",
		PlainHeading:          "New RSVP Received",
		PlainLabelEvent:       "Event",
		PlainLabelGuest:       "Guest",
		PlainLabelResponse:    "Response",
		PlainLabelEmail:       "Email",
		PlainLabelPhone:       "Phone",
		PlainLabelAdditional:  "Additional Guests",
		PlainFooterCTA:        "View your event dashboard:",
	},
	"de": {
		SubjectPrefix:         "Neue RSVP",
		Heading:               "Neue RSVP erhalten",
		IntroFormat:           "Jemand hat auf deine Veranstaltung <strong>%s</strong> geantwortet.",
		LabelGuest:            "Gast",
		LabelResponse:         "Antwort",
		LabelEmail:            "E-Mail",
		LabelPhone:            "Telefon",
		LabelAdditionalGuests: "Weitere Gäste",
		ChildrenSuffix:        "unter 12",
		ChildrenPlainFormat:   "%d Kind(er) unter 12",
		ButtonLabel:           "Zum Event-Dashboard",
		PlainHeading:          "Neue RSVP erhalten",
		PlainLabelEvent:       "Event",
		PlainLabelGuest:       "Gast",
		PlainLabelResponse:    "Antwort",
		PlainLabelEmail:       "E-Mail",
		PlainLabelPhone:       "Telefon",
		PlainLabelAdditional:  "Weitere Gäste",
		PlainFooterCTA:        "Zu deinem Event-Dashboard:",
	},
}

func organizerRSVPNotificationCopyFor(lang string) organizerRSVPNotificationCopy {
	if c, ok := organizerRSVPNotificationCopyByLang[lang]; ok {
		return c
	}
	return organizerRSVPNotificationCopyByLang[defaultLanguage]
}

// displayStatusAdmin returns a human-friendly, localized label for an RSVP
// status value, for the organizer notification email (distinct from the
// guest-facing displayStatusLocalized, since the organizer's language is
// independent of the event's guest language).
func displayStatusAdmin(lang, status string) string {
	labels := map[string]map[string]string{
		"en": {
			"attending":  "Attending",
			"maybe":      "Maybe",
			"declined":   "Can't make it",
			"pending":    "Pending",
			"waitlisted": "Waitlisted",
		},
		"de": {
			"attending":  "Zusage",
			"maybe":      "Vielleicht",
			"declined":   "Absage",
			"pending":    "Ausstehend",
			"waitlisted": "Warteliste",
		},
	}
	byLang, ok := labels[lang]
	if !ok {
		byLang = labels[defaultLanguage]
	}
	if label, ok := byLang[status]; ok {
		return label
	}
	return status
}

// --- Co-host invitation ---

type cohostInvitationCopy struct {
	Subject string
	Heading string
	// IntroFormat has two %s verbs: (HTML-escaped) added-by name, then event title.
	IntroFormat      string
	LabelEvent       string
	LabelDate        string
	LabelLocation    string
	ButtonLabel      string
	PlainIntroFormat string // same two verbs, no HTML
	PlainLabelEvent  string
	PlainLabelDate   string
	PlainLabelLoc    string
	PlainFooterCTA   string
}

var cohostInvitationCopyByLang = map[string]cohostInvitationCopy{
	"en": {
		Subject:          "You've been added as a co-host",
		Heading:          "You've Been Added as a Co-Host",
		IntroFormat:      "<strong>%s</strong> has added you as a co-host for <strong>%s</strong>. You can now manage RSVPs, send messages, and help run the event.",
		LabelEvent:       "Event",
		LabelDate:        "Date",
		LabelLocation:    "Location",
		ButtonLabel:      "View Event Dashboard",
		PlainIntroFormat: "%s has added you as a co-host for %s.",
		PlainLabelEvent:  "Event",
		PlainLabelDate:   "Date",
		PlainLabelLoc:    "Location",
		PlainFooterCTA:   "View the event dashboard:",
	},
	"de": {
		Subject:          "Du wurdest als Co-Host hinzugefügt",
		Heading:          "Du wurdest als Co-Host hinzugefügt",
		IntroFormat:      "<strong>%s</strong> hat dich als Co-Host für <strong>%s</strong> hinzugefügt. Du kannst jetzt RSVPs verwalten, Nachrichten senden und bei der Organisation helfen.",
		LabelEvent:       "Event",
		LabelDate:        "Datum",
		LabelLocation:    "Ort",
		ButtonLabel:      "Zum Event-Dashboard",
		PlainIntroFormat: "%s hat dich als Co-Host für %s hinzugefügt.",
		PlainLabelEvent:  "Event",
		PlainLabelDate:   "Datum",
		PlainLabelLoc:    "Ort",
		PlainFooterCTA:   "Zum Event-Dashboard:",
	},
}

func cohostInvitationCopyFor(lang string) cohostInvitationCopy {
	if c, ok := cohostInvitationCopyByLang[lang]; ok {
		return c
	}
	return cohostInvitationCopyByLang[defaultLanguage]
}

// --- Retention warning ---

type retentionWarningCopy struct {
	Subject           string
	Heading           string
	Intro             string
	LabelEvent        string
	LabelDeletionDate string
	Warning           string
	ButtonLabel       string
	FooterLine1       string
	FooterLine2       string
	PlainHeading      string
	// PlainIntroFormat has two %s verbs: event title, expiry date.
	PlainIntroFormat string
	PlainWarning     string
	PlainFooterCTA   string
}

var retentionWarningCopyByLang = map[string]retentionWarningCopy{
	"en": {
		Subject:           "Data Retention Notice",
		Heading:           "Data Retention Notice",
		Intro:             "Your event data is scheduled for automatic deletion soon. This is a courtesy reminder so you can take action if needed.",
		LabelEvent:        "Event",
		LabelDeletionDate: "Deletion Date",
		Warning:           "After this date, all event data including attendee RSVPs, messages, and invite cards will be permanently deleted. If you would like to keep this data, please log in and extend the retention period.",
		ButtonLabel:       "View Event",
		FooterLine1:       "This is an automated notice from OpenRSVP.",
		FooterLine2:       "You received this because you are the organizer of this event.",
		PlainHeading:      "Data Retention Notice",
		PlainIntroFormat:  "Your event \"%s\" is scheduled for automatic deletion on %s.",
		PlainWarning:      "After this date, all event data including attendee RSVPs, messages, and invite cards will be permanently deleted.",
		PlainFooterCTA:    "To extend the retention period, visit:",
	},
	"de": {
		Subject:           "Hinweis zur Datenaufbewahrung",
		Heading:           "Hinweis zur Datenaufbewahrung",
		Intro:             "Deine Veranstaltungsdaten werden bald automatisch gelöscht. Dies ist eine freundliche Erinnerung, damit du bei Bedarf handeln kannst.",
		LabelEvent:        "Event",
		LabelDeletionDate: "Löschdatum",
		Warning:           "Nach diesem Datum werden alle Veranstaltungsdaten, einschließlich Gäste-RSVPs, Nachrichten und Einladungskarten, dauerhaft gelöscht. Wenn du diese Daten behalten möchtest, melde dich bitte an und verlängere die Aufbewahrungsfrist.",
		ButtonLabel:       "Veranstaltung ansehen",
		FooterLine1:       "Dies ist ein automatischer Hinweis von OpenRSVP.",
		FooterLine2:       "Du erhältst diese Nachricht, weil du die veranstaltende Person bist.",
		PlainHeading:      "Hinweis zur Datenaufbewahrung",
		PlainIntroFormat:  "Deine Veranstaltung \"%s\" wird am %s automatisch gelöscht.",
		PlainWarning:      "Nach diesem Datum werden alle Veranstaltungsdaten, einschließlich Gäste-RSVPs, Nachrichten und Einladungskarten, dauerhaft gelöscht.",
		PlainFooterCTA:    "Um die Aufbewahrungsfrist zu verlängern, besuche:",
	},
}

func retentionWarningCopyFor(lang string) retentionWarningCopy {
	if c, ok := retentionWarningCopyByLang[lang]; ok {
		return c
	}
	return retentionWarningCopyByLang[defaultLanguage]
}

// --- Feedback confirmation ---

type feedbackConfirmationCopy struct {
	Subject string
	Heading string
	// IntroFormat has one %s verb for the (HTML-escaped, bold-wrapped) feedback type.
	IntroFormat      string
	FollowUpNote     string
	Closing          string
	PlainHeading     string
	PlainIntroFormat string // one %s verb, no HTML
	PlainFollowUp    string
	PlainClosing     string
}

var feedbackConfirmationCopyByLang = map[string]feedbackConfirmationCopy{
	"en": {
		Subject:          "We received your feedback — OpenRSVP",
		Heading:          "Thanks for your feedback!",
		IntroFormat:      "We received your <strong>%s</strong> submission and appreciate you taking the time to share it with us.",
		FollowUpNote:     "Since you opted in to follow-up contact, we may reach out to you at this email address if we have questions or updates related to your feedback.",
		Closing:          "Your feedback helps make OpenRSVP better for everyone.",
		PlainHeading:     "Thanks for your feedback!",
		PlainIntroFormat: "We received your %s submission and appreciate you taking the time to share it with us.",
		PlainFollowUp:    "Since you opted in to follow-up contact, we may reach out to you if we have questions or updates related to your feedback.",
		PlainClosing:     "Your feedback helps make OpenRSVP better for everyone.",
	},
	"de": {
		Subject:          "Wir haben dein Feedback erhalten — OpenRSVP",
		Heading:          "Danke für dein Feedback!",
		IntroFormat:      "Wir haben deine <strong>%s</strong>-Einreichung erhalten und schätzen, dass du dir die Zeit genommen hast, sie mit uns zu teilen.",
		FollowUpNote:     "Da du der Kontaktaufnahme zugestimmt hast, melden wir uns eventuell unter dieser E-Mail-Adresse bei dir, falls wir Fragen oder Neuigkeiten zu deinem Feedback haben.",
		Closing:          "Dein Feedback hilft dabei, OpenRSVP für alle besser zu machen.",
		PlainHeading:     "Danke für dein Feedback!",
		PlainIntroFormat: "Wir haben deine %s-Einreichung erhalten und schätzen, dass du dir die Zeit genommen hast, sie mit uns zu teilen.",
		PlainFollowUp:    "Da du der Kontaktaufnahme zugestimmt hast, melden wir uns eventuell bei dir, falls wir Fragen oder Neuigkeiten zu deinem Feedback haben.",
		PlainClosing:     "Dein Feedback hilft dabei, OpenRSVP für alle besser zu machen.",
	},
}

func feedbackConfirmationCopyFor(lang string) feedbackConfirmationCopy {
	if c, ok := feedbackConfirmationCopyByLang[lang]; ok {
		return c
	}
	return feedbackConfirmationCopyByLang[defaultLanguage]
}

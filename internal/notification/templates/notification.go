package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

//go:embed defaults/*.json
var defaultsFS embed.FS

//go:embed notification_envelope.html
var envelopeFS embed.FS

var envelopeTmpl *template.Template

// TemplateDefault is a default subject/body pair for a guest notification
// message type in a given language.
type TemplateDefault struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// defaultLanguage is used whenever an event's language is unsupported or a
// default is missing for the requested language.
const defaultLanguage = "en"

// defaultsByType holds default templates keyed by [messageType][lang].
var defaultsByType map[string]map[string]TemplateDefault

func init() {
	envelopeTmpl = template.Must(template.ParseFS(envelopeFS, "notification_envelope.html"))

	entries, err := defaultsFS.ReadDir("defaults")
	if err != nil {
		panic(fmt.Sprintf("read embedded template defaults: %v", err))
	}

	defaultsByType = make(map[string]map[string]TemplateDefault)
	for _, entry := range entries {
		name := entry.Name() // e.g. "rsvp_confirmation.en.json"
		parts := strings.SplitN(strings.TrimSuffix(name, ".json"), ".", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintf("unexpected template default filename: %s", name))
		}
		messageType, lang := parts[0], parts[1]

		data, err := defaultsFS.ReadFile("defaults/" + name)
		if err != nil {
			panic(fmt.Sprintf("read embedded template default %s: %v", name, err))
		}

		var def TemplateDefault
		if err := json.Unmarshal(data, &def); err != nil {
			panic(fmt.Sprintf("parse embedded template default %s: %v", name, err))
		}

		if defaultsByType[messageType] == nil {
			defaultsByType[messageType] = make(map[string]TemplateDefault)
		}
		defaultsByType[messageType][lang] = def
	}
}

// DefaultFor returns the default subject/body template for a message type in
// the given language, falling back to defaultLanguage if the requested
// language has no default for that type.
func DefaultFor(messageType, lang string) TemplateDefault {
	byLang := defaultsByType[messageType]
	if def, ok := byLang[lang]; ok {
		return def
	}
	return byLang[defaultLanguage]
}

// Interpolate replaces {key} placeholders in text with the given raw values.
// Placeholders with no matching key are left as literal text so a typo in an
// organizer-authored template doesn't break the send.
func Interpolate(text string, vars map[string]string) string {
	if len(vars) == 0 {
		return text
	}
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(text)
}

// envelopeChrome holds the small set of fixed, translated strings that
// surround every guest notification's envelope (not organizer-editable).
type envelopeChrome struct {
	HelperText string
	FooterText string
}

var envelopeChromeByLang = map[string]envelopeChrome{
	"en": {
		HelperText: "If the button does not work, copy and paste this link into your browser:",
		FooterText: "© OpenRSVP — Simple event RSVPs",
	},
	"de": {
		HelperText: "Falls der Button nicht funktioniert, kopiere diesen Link in deinen Browser:",
		FooterText: "© OpenRSVP — Einfache Veranstaltungs-RSVPs",
	},
}

func envelopeChromeFor(lang string) envelopeChrome {
	if c, ok := envelopeChromeByLang[lang]; ok {
		return c
	}
	return envelopeChromeByLang[defaultLanguage]
}

// displayStatusLocalized returns a human-friendly, language-appropriate label
// for an RSVP status value, used when interpolating guest-facing templates.
func displayStatusLocalized(lang, status string) string {
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

// germanMonths maps time.Month to its German name, used by FormatEventDate.
// Go's time.Format layouts (e.g. "January 2, 2006") always render English
// month/weekday names -- the stdlib has no locale support -- so localized
// date strings need to be built by hand.
var germanMonths = map[time.Month]string{
	time.January:   "Januar",
	time.February:  "Februar",
	time.March:     "März",
	time.April:     "April",
	time.May:       "Mai",
	time.June:      "Juni",
	time.July:      "Juli",
	time.August:    "August",
	time.September: "September",
	time.October:   "Oktober",
	time.November:  "November",
	time.December:  "Dezember",
}

// InTimezone converts a stored (UTC) event time to the event's display
// timezone, so notification bodies show the same local time the organizer
// entered rather than the raw UTC hour. Falls back to UTC if tz is empty or
// unrecognized, mirroring internal/calendar.GenerateICS's handling.
func InTimezone(t time.Time, tz string) time.Time {
	if tz == "" {
		return t.UTC()
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return t.UTC()
	}
	return t.In(loc)
}

// FormatEventDate formats a time for display inside a guest notification, in
// the given language. Times are expected to already be in the event's
// display timezone -- pass the result of InTimezone, not the raw stored UTC
// value.
func FormatEventDate(t time.Time, lang string) string {
	switch lang {
	case "de":
		return fmt.Sprintf("%d. %s %d um %02d:%02d Uhr", t.Day(), germanMonths[t.Month()], t.Year(), t.Hour(), t.Minute())
	default:
		return t.Format("January 2, 2006 at 3:04 PM")
	}
}

// DisplayStatusLocalized is the exported form of displayStatusLocalized, used
// by callers outside this package to build the rsvpStatus template variable.
func DisplayStatusLocalized(lang, status string) string {
	return displayStatusLocalized(lang, status)
}

// ctaLabelsByType holds the call-to-action button label for each guest
// notification message type, per language.
var ctaLabelsByType = map[string]map[string]string{
	"rsvp_confirmation": {
		"en": "Modify RSVP",
		"de": "RSVP ändern",
	},
	"cancellation": {
		"en": "View Event",
		"de": "Veranstaltung ansehen",
	},
	"reminder": {
		"en": "View Invitation",
		"de": "Einladung ansehen",
	},
	"waitlist_promotion": {
		"en": "View RSVP",
		"de": "RSVP ansehen",
	},
	"import_invite": {
		"en": "View Invitation",
		"de": "Einladung ansehen",
	},
}

// CTALabel returns the call-to-action button label for a guest notification
// message type in the given language, falling back to defaultLanguage.
func CTALabel(messageType, lang string) string {
	byLang, ok := ctaLabelsByType[messageType]
	if !ok {
		return ""
	}
	if label, ok := byLang[lang]; ok {
		return label
	}
	return byLang[defaultLanguage]
}

// notificationEnvelopeData holds the template data for the shared guest
// notification envelope.
type notificationEnvelopeData struct {
	Lang       string
	Subject    string
	WordMark   string
	BodyHTML   template.HTML
	CTAURL     string
	CTALabel   string
	HelperText string
	FooterText string
	Colors     EmailColors
}

// RenderNotification renders a guest-facing notification into the shared
// branded HTML envelope and a plain-text counterpart. bodyTemplate is the
// raw (uninterpolated) subject/body template text (either an organizer
// override or a language default); vars holds the raw variable values to
// interpolate (e.g. guestName, eventTitle). ctaURL/ctaLabel add an optional
// call-to-action button (e.g. "Modify RSVP" linking to the RSVP page).
func RenderNotification(lang, subjectTemplate, bodyTemplate string, vars map[string]string, ctaURL, ctaLabel string) (subject, html, plain string, err error) {
	subject = Interpolate(subjectTemplate, vars)
	plain = Interpolate(bodyTemplate, vars)
	if ctaURL != "" {
		plain = plain + "\n\n" + ctaLabel + ":\n" + ctaURL
	}

	escapedVars := make(map[string]string, len(vars))
	for k, v := range vars {
		escapedVars[k] = template.HTMLEscapeString(v)
	}
	bodyHTML := Interpolate(template.HTMLEscapeString(bodyTemplate), escapedVars)
	bodyHTML = strings.ReplaceAll(bodyHTML, "\n", "<br>")

	chrome := envelopeChromeFor(lang)
	data := notificationEnvelopeData{
		Lang:       lang,
		Subject:    subject,
		WordMark:   "OpenRSVP",
		BodyHTML:   template.HTML(bodyHTML), //nolint:gosec // bodyHTML was built from escaped, placeholder-substituted text above
		CTAURL:     ctaURL,
		CTALabel:   ctaLabel,
		HelperText: chrome.HelperText,
		FooterText: chrome.FooterText,
		Colors:     DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := envelopeTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render notification envelope: %w", err)
	}

	return subject, buf.String(), plain, nil
}

// SMSFrom derives a short SMS body from an already-interpolated plain-text
// notification body and a trailing link, truncating the body (never the
// link) to a safe length so the link always stays intact -- a guest should
// never receive a confirmation SMS whose management link got cut off.
func SMSFrom(body, link string, maxLen int) string {
	full := body
	if link != "" {
		full = body + "\n\n" + link
	}
	if len(full) <= maxLen {
		return full
	}
	if link == "" {
		truncated := body[:maxLen]
		if idx := strings.LastIndex(truncated, " "); idx > 0 {
			truncated = truncated[:idx]
		}
		return truncated + "…"
	}
	suffix := "…\n\n" + link
	available := maxLen - len(suffix)
	if available < 0 {
		available = 0
	}
	truncatedBody := body
	if len(truncatedBody) > available {
		truncatedBody = truncatedBody[:available]
		if idx := strings.LastIndex(truncatedBody, " "); idx > 0 {
			truncatedBody = truncatedBody[:idx]
		}
	}
	return truncatedBody + suffix
}

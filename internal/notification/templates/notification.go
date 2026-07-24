package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
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
// notification body, truncating it to a safe length while keeping the
// trailing link (if present) intact.
func SMSFrom(plain string, maxLen int) string {
	if len(plain) <= maxLen {
		return plain
	}
	truncated := plain[:maxLen]
	if idx := strings.LastIndex(truncated, " "); idx > 0 {
		truncated = truncated[:idx]
	}
	return truncated + "…"
}

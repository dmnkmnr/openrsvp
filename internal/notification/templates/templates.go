package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed magic_link.html event_reminder.html retention_warning.html organizer_rsvp_notification.html feedback_confirmation.html rsvp_lookup.html cohost_invitation.html guest_message_notification.html
var templateFS embed.FS

var (
	magicLinkTmpl            *template.Template
	eventReminderTmpl        *template.Template
	retentionWarningTmpl     *template.Template
	organizerRSVPNotifyTmpl  *template.Template
	feedbackConfirmationTmpl *template.Template
	rsvpLookupTmpl           *template.Template
	cohostInvitationTmpl     *template.Template
	guestMessageNotifyTmpl   *template.Template
)

func init() {
	magicLinkTmpl = template.Must(template.ParseFS(templateFS, "magic_link.html"))
	eventReminderTmpl = template.Must(template.ParseFS(templateFS, "event_reminder.html"))
	retentionWarningTmpl = template.Must(template.ParseFS(templateFS, "retention_warning.html"))
	organizerRSVPNotifyTmpl = template.Must(template.ParseFS(templateFS, "organizer_rsvp_notification.html"))
	feedbackConfirmationTmpl = template.Must(template.ParseFS(templateFS, "feedback_confirmation.html"))
	rsvpLookupTmpl = template.Must(template.ParseFS(templateFS, "rsvp_lookup.html"))
	cohostInvitationTmpl = template.Must(template.ParseFS(templateFS, "cohost_invitation.html"))
	guestMessageNotifyTmpl = template.Must(template.ParseFS(templateFS, "guest_message_notification.html"))
}

// magicLinkData holds the template data for a magic link email.
type magicLinkData struct {
	URL           string
	ExpiryMinutes int
	Heading       string
	Intro         string
	ButtonLabel   string
	HelperText    string
	ExpiryNote    string
	FooterText    string
	Colors        EmailColors
}

// eventReminderData holds the template data for an event reminder email.
type eventReminderData struct {
	EventTitle string
	EventDate  string
	Location   string
	Message    string
	InviteURL  string
	Colors     EmailColors
}

// retentionWarningData holds the template data for a retention warning email.
type retentionWarningData struct {
	EventTitle        string
	ExpiresAt         string
	DashboardURL      string
	Heading           string
	Intro             string
	LabelEvent        string
	LabelDeletionDate string
	Warning           string
	ButtonLabel       string
	FooterLine1       string
	FooterLine2       string
	Colors            EmailColors
}

// organizerRSVPNotificationData holds the template data for notifying an
// organizer about a new or updated RSVP.
type organizerRSVPNotificationData struct {
	EventTitle            string
	GuestName             string
	RSVPStatus            string
	GuestEmail            string
	GuestPhone            string
	PlusOnes              int
	PlusOnesChildren      int
	DashboardURL          string
	Heading               string
	IntroHTML             template.HTML
	LabelGuest            string
	LabelResponse         string
	LabelEmail            string
	LabelPhone            string
	LabelAdditionalGuests string
	ChildrenSuffix        string
	ButtonLabel           string
	FooterText            string
	Colors                EmailColors
}

// RenderMagicLink renders the magic link email template and returns the
// localized subject, the HTML body, and a plain text fallback.
func RenderMagicLink(lang, baseURL, token string, expiryMinutes int) (subject, html, plain string, err error) {
	url := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, token)
	msgCopy := magicLinkCopyFor(lang)
	chrome := envelopeChromeFor(lang)
	expiryNote := fmt.Sprintf(msgCopy.ExpiryNote, expiryMinutes)

	data := magicLinkData{
		URL:           url,
		ExpiryMinutes: expiryMinutes,
		Heading:       msgCopy.Heading,
		Intro:         msgCopy.Intro,
		ButtonLabel:   msgCopy.ButtonLabel,
		HelperText:    chrome.HelperText,
		ExpiryNote:    expiryNote,
		FooterText:    chrome.FooterText,
		Colors:        DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := magicLinkTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render magic link template: %w", err)
	}

	plainText := fmt.Sprintf(
		"%s\n\n%s\n%s\n\n%s",
		msgCopy.Heading, msgCopy.Intro, url, expiryNote,
	)

	return msgCopy.Subject, buf.String(), plainText, nil
}

// RenderEventReminder renders the event reminder email template and returns
// the HTML body and a plain text fallback.
func RenderEventReminder(eventTitle, eventDate, location, message, inviteURL string) (html, plain string, err error) {
	data := eventReminderData{
		EventTitle: eventTitle,
		EventDate:  eventDate,
		Location:   location,
		Message:    message,
		InviteURL:  inviteURL,
		Colors:     DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := eventReminderTmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("render event reminder template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Event Reminder\n\n")
	sb.WriteString(fmt.Sprintf("Event: %s\n", eventTitle))
	sb.WriteString(fmt.Sprintf("Date: %s\n", eventDate))
	sb.WriteString(fmt.Sprintf("Location: %s\n\n", location))
	if message != "" {
		sb.WriteString(fmt.Sprintf("Message from the organizer:\n%s\n\n", message))
	}
	sb.WriteString(fmt.Sprintf("View your invitation:\n%s\n", inviteURL))

	return buf.String(), sb.String(), nil
}

// RenderRetentionWarning renders the retention warning email template and
// returns the localized subject, the HTML body, and a plain text fallback.
func RenderRetentionWarning(lang, eventTitle, expiresAt, dashboardURL string) (subject, html, plain string, err error) {
	msgCopy := retentionWarningCopyFor(lang)
	data := retentionWarningData{
		EventTitle:        eventTitle,
		ExpiresAt:         expiresAt,
		DashboardURL:      dashboardURL,
		Heading:           msgCopy.Heading,
		Intro:             msgCopy.Intro,
		LabelEvent:        msgCopy.LabelEvent,
		LabelDeletionDate: msgCopy.LabelDeletionDate,
		Warning:           msgCopy.Warning,
		ButtonLabel:       msgCopy.ButtonLabel,
		FooterLine1:       msgCopy.FooterLine1,
		FooterLine2:       msgCopy.FooterLine2,
		Colors:            DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := retentionWarningTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render retention warning template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(msgCopy.PlainHeading + "\n\n")
	sb.WriteString(fmt.Sprintf(msgCopy.PlainIntroFormat, eventTitle, expiresAt) + "\n\n")
	sb.WriteString(msgCopy.PlainWarning + "\n\n")
	if dashboardURL != "" {
		sb.WriteString(fmt.Sprintf("%s\n%s\n", msgCopy.PlainFooterCTA, dashboardURL))
	}

	return msgCopy.Subject, buf.String(), sb.String(), nil
}

// RenderOrganizerRSVPNotification renders the organizer RSVP notification
// email and returns the localized subject, the HTML body, and a plain text
// fallback.
func RenderOrganizerRSVPNotification(lang, eventTitle, guestName, rsvpStatus, guestEmail, guestPhone string, plusOnes, plusOnesChildren int, dashboardURL string) (subject, html, plain string, err error) {
	msgCopy := organizerRSVPNotificationCopyFor(lang)
	chrome := envelopeChromeFor(lang)
	label := displayStatusAdmin(lang, rsvpStatus)

	data := organizerRSVPNotificationData{
		EventTitle:            eventTitle,
		GuestName:             guestName,
		RSVPStatus:            label,
		GuestEmail:            guestEmail,
		GuestPhone:            guestPhone,
		PlusOnes:              plusOnes,
		PlusOnesChildren:      plusOnesChildren,
		DashboardURL:          dashboardURL,
		Heading:               msgCopy.Heading,
		IntroHTML:             template.HTML(fmt.Sprintf(msgCopy.IntroFormat, template.HTMLEscapeString(eventTitle))), //nolint:gosec // eventTitle is HTML-escaped above
		LabelGuest:            msgCopy.LabelGuest,
		LabelResponse:         msgCopy.LabelResponse,
		LabelEmail:            msgCopy.LabelEmail,
		LabelPhone:            msgCopy.LabelPhone,
		LabelAdditionalGuests: msgCopy.LabelAdditionalGuests,
		ChildrenSuffix:        msgCopy.ChildrenSuffix,
		ButtonLabel:           msgCopy.ButtonLabel,
		FooterText:            chrome.FooterText,
		Colors:                DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := organizerRSVPNotifyTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render organizer rsvp notification template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(msgCopy.PlainHeading + "\n\n")
	sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelEvent, eventTitle))
	sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelGuest, guestName))
	sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelResponse, label))
	if guestEmail != "" {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelEmail, guestEmail))
	}
	if guestPhone != "" {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelPhone, guestPhone))
	}
	if plusOnes > 0 {
		if plusOnesChildren > 0 {
			sb.WriteString(fmt.Sprintf("%s: +%d (%s)\n", msgCopy.PlainLabelAdditional, plusOnes, fmt.Sprintf(msgCopy.ChildrenPlainFormat, plusOnesChildren)))
		} else {
			sb.WriteString(fmt.Sprintf("%s: +%d\n", msgCopy.PlainLabelAdditional, plusOnes))
		}
	}
	sb.WriteString(fmt.Sprintf("\n%s\n%s\n", msgCopy.PlainFooterCTA, dashboardURL))

	return msgCopy.SubjectPrefix, buf.String(), sb.String(), nil
}

// rsvpLookupData holds the template data for an RSVP lookup email.
type rsvpLookupData struct {
	EventTitle string
	ModifyURL  string
	Colors     EmailColors
}

// RenderRSVPLookup renders the RSVP lookup magic link email template and
// returns the HTML body and a plain text fallback.
func RenderRSVPLookup(eventTitle, modifyURL string) (html, plain string, err error) {
	data := rsvpLookupData{
		EventTitle: eventTitle,
		ModifyURL:  modifyURL,
		Colors:     DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := rsvpLookupTmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("render rsvp lookup template: %w", err)
	}

	plainText := fmt.Sprintf(
		"Find Your RSVP\n\nClick the link below to view and manage your RSVP for %s:\n%s\n\nThis link is personal — please don't share it.",
		eventTitle, modifyURL,
	)

	return buf.String(), plainText, nil
}

// feedbackConfirmationData holds the template data for a feedback confirmation email.
type feedbackConfirmationData struct {
	FeedbackType  string
	AllowFollowUp bool
	Heading       string
	IntroHTML     template.HTML
	FollowUpNote  string
	Closing       string
	FooterText    string
	Colors        EmailColors
}

// RenderFeedbackConfirmation renders the feedback confirmation email template
// and returns the localized subject, the HTML body, and a plain text fallback.
func RenderFeedbackConfirmation(feedbackType string, allowFollowUp bool, lang string) (subject, htmlBody, plain string, err error) {
	msgCopy := feedbackConfirmationCopyFor(lang)
	chrome := envelopeChromeFor(lang)

	data := feedbackConfirmationData{
		FeedbackType:  feedbackType,
		AllowFollowUp: allowFollowUp,
		Heading:       msgCopy.Heading,
		IntroHTML:     template.HTML(fmt.Sprintf(msgCopy.IntroFormat, template.HTMLEscapeString(feedbackType))), //nolint:gosec // feedbackType is HTML-escaped above
		FollowUpNote:  msgCopy.FollowUpNote,
		Closing:       msgCopy.Closing,
		FooterText:    chrome.FooterText,
		Colors:        DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := feedbackConfirmationTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render feedback confirmation template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(msgCopy.PlainHeading + "\n\n")
	sb.WriteString(fmt.Sprintf(msgCopy.PlainIntroFormat, feedbackType) + "\n\n")
	if allowFollowUp {
		sb.WriteString(msgCopy.PlainFollowUp + "\n\n")
	}
	sb.WriteString(msgCopy.PlainClosing + "\n")

	return msgCopy.Subject, buf.String(), sb.String(), nil
}

// cohostInvitationData holds the template data for a co-host invitation email.
type cohostInvitationData struct {
	EventTitle    string
	EventDate     string
	Location      string
	AddedByName   string
	DashboardURL  string
	Heading       string
	IntroHTML     template.HTML
	LabelEvent    string
	LabelDate     string
	LabelLocation string
	ButtonLabel   string
	HelperText    string
	FooterText    string
	Colors        EmailColors
}

// RenderCoHostInvitation renders the co-host invitation email template and
// returns the localized subject, the HTML body, and a plain text fallback.
// lang is the invited co-host's (the recipient's) language, not the inviting
// organizer's.
func RenderCoHostInvitation(lang, eventTitle, eventDate, location, addedByName, dashboardURL string) (subject, html, plain string, err error) {
	msgCopy := cohostInvitationCopyFor(lang)
	chrome := envelopeChromeFor(lang)

	data := cohostInvitationData{
		EventTitle:    eventTitle,
		EventDate:     eventDate,
		Location:      location,
		AddedByName:   addedByName,
		DashboardURL:  dashboardURL,
		Heading:       msgCopy.Heading,
		IntroHTML:     template.HTML(fmt.Sprintf(msgCopy.IntroFormat, template.HTMLEscapeString(addedByName), template.HTMLEscapeString(eventTitle))), //nolint:gosec // both values are HTML-escaped above
		LabelEvent:    msgCopy.LabelEvent,
		LabelDate:     msgCopy.LabelDate,
		LabelLocation: msgCopy.LabelLocation,
		ButtonLabel:   msgCopy.ButtonLabel,
		HelperText:    chrome.HelperText,
		FooterText:    chrome.FooterText,
		Colors:        DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := cohostInvitationTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render cohost invitation template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(msgCopy.Heading + "\n\n")
	sb.WriteString(fmt.Sprintf(msgCopy.PlainIntroFormat, addedByName, eventTitle) + "\n\n")
	sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelEvent, eventTitle))
	sb.WriteString(fmt.Sprintf("%s: %s\n", msgCopy.PlainLabelDate, eventDate))
	sb.WriteString(fmt.Sprintf("%s: %s\n\n", msgCopy.PlainLabelLoc, location))
	sb.WriteString(fmt.Sprintf("%s\n%s\n", msgCopy.PlainFooterCTA, dashboardURL))

	return msgCopy.Subject, buf.String(), sb.String(), nil
}

// guestMessageNotificationData holds the template data for the "a guest sent
// you a message" notification email.
type guestMessageNotificationData struct {
	DashboardURL    string
	Heading         string
	IntroHTML       template.HTML
	LabelMessage    string
	MessageBodyHTML template.HTML
	ButtonLabel     string
	HelperText      string
	FooterText      string
	Colors          EmailColors
}

// RenderGuestMessageNotification renders the "a guest sent you a message"
// organizer notification and returns the localized subject, the HTML body,
// and a plain text fallback. lang is the organizer's own account language.
func RenderGuestMessageNotification(lang, eventTitle, guestName, messageSubject, messageBody, dashboardURL string) (subject, html, plain string, err error) {
	msgCopy := guestMessageNotificationCopyFor(lang)
	chrome := envelopeChromeFor(lang)

	escapedBody := template.HTMLEscapeString(messageBody)
	bodyHTML := strings.ReplaceAll(escapedBody, "\n", "<br>")

	data := guestMessageNotificationData{
		DashboardURL:    dashboardURL,
		Heading:         msgCopy.Heading,
		IntroHTML:       template.HTML(fmt.Sprintf(msgCopy.IntroFormat, template.HTMLEscapeString(guestName), template.HTMLEscapeString(eventTitle))), //nolint:gosec // both values are HTML-escaped above
		LabelMessage:    msgCopy.LabelMessage,
		MessageBodyHTML: template.HTML(bodyHTML), //nolint:gosec // messageBody was HTML-escaped above before the newline-to-<br> substitution
		ButtonLabel:     msgCopy.ButtonLabel,
		HelperText:      chrome.HelperText,
		FooterText:      chrome.FooterText,
		Colors:          DefaultEmailColors(),
	}

	var buf bytes.Buffer
	if err := guestMessageNotifyTmpl.Execute(&buf, data); err != nil {
		return "", "", "", fmt.Errorf("render guest message notification template: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(msgCopy.Heading + "\n\n")
	sb.WriteString(fmt.Sprintf(msgCopy.PlainIntroFormat, guestName, eventTitle) + "\n\n")
	sb.WriteString(fmt.Sprintf("%s: %s\n\n", msgCopy.PlainLabelMessage, messageBody))
	sb.WriteString(fmt.Sprintf("%s\n%s\n", msgCopy.PlainFooterCTA, dashboardURL))

	subject = fmt.Sprintf("%s %s — %s", msgCopy.SubjectPrefix, guestName, messageSubject)

	return subject, buf.String(), sb.String(), nil
}

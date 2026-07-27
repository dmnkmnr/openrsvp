package templates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderRetentionWarning(t *testing.T) {
	subject, html, plain, err := RenderRetentionWarning("en", "Birthday Party", "March 15, 2026", "http://localhost:8080/events")
	require.NoError(t, err)

	assert.Equal(t, "Data Retention Notice", subject)
	assert.Contains(t, html, "Birthday Party")
	assert.Contains(t, html, "March 15, 2026")
	assert.Contains(t, html, "http://localhost:8080/events")
	assert.Contains(t, html, "Data Retention Notice")

	assert.Contains(t, plain, "Birthday Party")
	assert.Contains(t, plain, "March 15, 2026")
	assert.Contains(t, plain, "http://localhost:8080/events")
	assert.Contains(t, plain, "permanently deleted")
}

func TestRenderRetentionWarningNoDashboardURL(t *testing.T) {
	_, html, plain, err := RenderRetentionWarning("en", "Garden Party", "April 20, 2026", "")
	require.NoError(t, err)

	assert.Contains(t, html, "Garden Party")
	assert.Contains(t, html, "April 20, 2026")
	assert.NotContains(t, html, "View Event")

	assert.Contains(t, plain, "Garden Party")
	assert.NotContains(t, plain, "visit:")
}

func TestRenderRetentionWarningGerman(t *testing.T) {
	subject, html, plain, err := RenderRetentionWarning("de", "Geburtstagsfeier", "15. März 2026", "http://localhost:8080/events")
	require.NoError(t, err)

	assert.Equal(t, "Hinweis zur Datenaufbewahrung", subject)
	assert.Contains(t, html, "Geburtstagsfeier")
	assert.Contains(t, html, "Hinweis zur Datenaufbewahrung")
	assert.Contains(t, plain, "dauerhaft gelöscht")
}

func TestRenderMagicLink(t *testing.T) {
	subject, html, plain, err := RenderMagicLink("en", "http://localhost:8080", "abc123token", 15)
	require.NoError(t, err)

	assert.Equal(t, "Sign in to OpenRSVP", subject)
	assert.Contains(t, html, "abc123token")
	assert.Contains(t, html, "http://localhost:8080")
	assert.Contains(t, plain, "15 minutes")
}

func TestRenderMagicLinkGerman(t *testing.T) {
	subject, html, plain, err := RenderMagicLink("de", "http://localhost:8080", "abc123token", 15)
	require.NoError(t, err)

	assert.Equal(t, "Anmeldung bei OpenRSVP", subject)
	assert.Contains(t, html, "abc123token")
	assert.Contains(t, plain, "15 Minuten")
}

func TestRenderOrganizerRSVPNotification(t *testing.T) {
	subject, html, plain, err := RenderOrganizerRSVPNotification("en", "Pool Party", "Alice", "attending", "alice@example.com", "", 2, 1, "http://localhost/events/1")
	require.NoError(t, err)

	assert.Equal(t, "New RSVP", subject)
	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, html, "Alice")
	assert.Contains(t, html, "Attending")
	assert.Contains(t, html, "under 12")
	assert.Contains(t, plain, "Alice")
}

func TestRenderOrganizerRSVPNotificationGerman(t *testing.T) {
	subject, html, plain, err := RenderOrganizerRSVPNotification("de", "Poolparty", "Alice", "attending", "", "", 0, 0, "http://localhost/events/1")
	require.NoError(t, err)

	assert.Equal(t, "Neue RSVP", subject)
	assert.Contains(t, html, "Zusage")
	assert.Contains(t, plain, "Zusage")
}

func TestRenderCoHostInvitation(t *testing.T) {
	subject, html, plain, err := RenderCoHostInvitation("en", "Pool Party", "June 5, 2026", "Backyard", "Bob", "http://localhost/events/1")
	require.NoError(t, err)

	assert.Equal(t, "You've been added as a co-host", subject)
	assert.Contains(t, html, "Bob")
	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, plain, "Bob")
}

func TestRenderCoHostInvitationGerman(t *testing.T) {
	subject, html, plain, err := RenderCoHostInvitation("de", "Poolparty", "5. Juni 2026", "Garten", "Bob", "http://localhost/events/1")
	require.NoError(t, err)

	assert.Equal(t, "Du wurdest als Co-Host hinzugefügt", subject)
	assert.Contains(t, html, "Co-Host")
	assert.Contains(t, plain, "Co-Host")
}

func TestRenderGuestMessageNotification(t *testing.T) {
	subject, html, plain, err := RenderGuestMessageNotification(
		"en", "Pool Party", "Alice", "Question about parking", "Is there parking nearby?",
		"http://localhost/events/1/messages",
	)
	require.NoError(t, err)

	assert.Equal(t, "New message from Alice — Question about parking", subject)
	assert.Contains(t, html, "New Message From a Guest")
	assert.Contains(t, html, "Alice")
	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, html, "Is there parking nearby?")
	assert.NotContains(t, html, "Reminder")
	assert.Contains(t, plain, "Alice")
	assert.Contains(t, plain, "Is there parking nearby?")
}

func TestRenderGuestMessageNotificationGerman(t *testing.T) {
	subject, html, plain, err := RenderGuestMessageNotification(
		"de", "Poolparty", "Alice", "Frage zum Parken", "Gibt es Parkplätze in der Nähe?",
		"http://localhost/events/1/messages",
	)
	require.NoError(t, err)

	assert.Equal(t, "Neue Nachricht von Alice — Frage zum Parken", subject)
	assert.Contains(t, html, "Neue Nachricht von einem Gast")
	assert.Contains(t, html, "Gibt es Parkplätze in der Nähe?")
	assert.Contains(t, plain, "Gibt es Parkplätze in der Nähe?")
}

func TestRenderRSVPLookup(t *testing.T) {
	subject, html, plain, err := RenderRSVPLookup("en", "Pool Party", "http://localhost/r/tok")
	require.NoError(t, err)

	assert.Equal(t, "Your RSVP Link — Pool Party", subject)
	assert.Contains(t, html, "Find Your RSVP")
	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, html, "View My RSVP")
	assert.Contains(t, plain, "Pool Party")
}

func TestRenderRSVPLookupGerman(t *testing.T) {
	subject, html, plain, err := RenderRSVPLookup("de", "Poolparty", "http://localhost/r/tok")
	require.NoError(t, err)

	assert.Equal(t, "Dein RSVP-Link — Poolparty", subject)
	assert.Contains(t, html, "Deine RSVP finden")
	assert.Contains(t, html, "Poolparty")
	assert.Contains(t, html, "Meine RSVP ansehen")
	assert.NotContains(t, html, "Find Your RSVP")
	assert.Contains(t, plain, "Poolparty")
}

func TestRenderFeedbackConfirmation(t *testing.T) {
	subject, html, plain, err := RenderFeedbackConfirmation("bug", true, "en")
	require.NoError(t, err)

	assert.Equal(t, "We received your feedback — OpenRSVP", subject)
	assert.Contains(t, html, "bug")
	assert.Contains(t, plain, "bug")
}

func TestRenderFeedbackConfirmationGerman(t *testing.T) {
	subject, html, plain, err := RenderFeedbackConfirmation("bug", false, "de")
	require.NoError(t, err)

	assert.Equal(t, "Wir haben dein Feedback erhalten — OpenRSVP", subject)
	assert.Contains(t, html, "Danke für dein Feedback")
	assert.Contains(t, plain, "Danke für dein Feedback")
}

func TestRenderEventReminder(t *testing.T) {
	html, plain, err := RenderEventReminder("Pool Party", "June 5, 2026", "Backyard", "Remember!", "http://localhost/i/xyz")
	require.NoError(t, err)

	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, html, "Remember!")
	assert.Contains(t, plain, "Pool Party")
	assert.Contains(t, plain, "Remember!")
	assert.Contains(t, plain, "http://localhost/i/xyz")
}

func TestRenderEventReminderNoMessage(t *testing.T) {
	html, plain, err := RenderEventReminder("Quick Event", "July 1, 2026", "Park", "", "http://localhost/i/abc")
	require.NoError(t, err)

	assert.Contains(t, html, "Quick Event")
	assert.NotContains(t, html, "Message from the organizer")
	assert.NotContains(t, plain, "Message from the organizer")
}

func TestInterpolate(t *testing.T) {
	out := Interpolate("Hi {guestName}, welcome to {eventTitle}!", map[string]string{
		"guestName":  "Alex",
		"eventTitle": "Pool Party",
	})
	assert.Equal(t, "Hi Alex, welcome to Pool Party!", out)
}

func TestInterpolateUnknownPlaceholderLeftLiteral(t *testing.T) {
	out := Interpolate("Hi {guestName}, code: {unknownVar}", map[string]string{"guestName": "Alex"})
	assert.Equal(t, "Hi Alex, code: {unknownVar}", out)
}

func TestDefaultForKnownAndFallback(t *testing.T) {
	en := DefaultFor("rsvp_confirmation", "en")
	assert.NotEmpty(t, en.Subject)
	assert.NotEmpty(t, en.Body)

	de := DefaultFor("rsvp_confirmation", "de")
	assert.NotEmpty(t, de.Subject)
	assert.NotEqual(t, en.Subject, de.Subject)

	// Unsupported language falls back to English.
	fallback := DefaultFor("rsvp_confirmation", "fr")
	assert.Equal(t, en.Subject, fallback.Subject)
}

func TestRenderNotification(t *testing.T) {
	def := DefaultFor("rsvp_confirmation", "en")
	vars := map[string]string{
		"guestName":  "Alex",
		"eventTitle": "Pool Party",
		"eventDate":  "June 5, 2026",
		"location":   "Backyard",
		"rsvpStatus": "Attending",
	}

	subject, html, plain, err := RenderNotification("en", def.Subject, def.Body, vars, "http://localhost/r/tok", CTALabel("rsvp_confirmation", "en"))
	require.NoError(t, err)

	assert.Contains(t, subject, "Pool Party")
	assert.Contains(t, html, "Alex")
	assert.Contains(t, html, "Pool Party")
	assert.Contains(t, html, "http://localhost/r/tok")
	assert.Contains(t, plain, "Alex")
	assert.Contains(t, plain, "http://localhost/r/tok")
}

func TestRenderNotificationEscapesHTML(t *testing.T) {
	subject, html, plain, err := RenderNotification("en", "Subject", "Hi {guestName}", map[string]string{
		"guestName": "<script>alert(1)</script>",
	}, "", "")
	require.NoError(t, err)
	_ = subject

	assert.NotContains(t, html, "<script>")
	assert.Contains(t, html, "&lt;script&gt;")
	assert.Contains(t, plain, "<script>alert(1)</script>")
}

func TestSMSFromTruncates(t *testing.T) {
	long := "This is a very long reminder message that definitely exceeds the short SMS length limit we configured for this test case here"
	out := SMSFrom(long, 40)
	assert.LessOrEqual(t, len(out), 44) // 40 + ellipsis bytes
	assert.True(t, len(out) < len(long))
}

func TestSMSFromNoTruncationNeeded(t *testing.T) {
	short := "Short message"
	assert.Equal(t, short, SMSFrom(short, 40))
}

func TestFormatEventDateEnglish(t *testing.T) {
	d := time.Date(2026, time.June, 5, 15, 4, 0, 0, time.UTC)
	assert.Equal(t, "June 5, 2026 at 3:04 PM", FormatEventDate(d, "en"))
}

func TestFormatEventDateGerman(t *testing.T) {
	d := time.Date(2026, time.June, 5, 15, 4, 0, 0, time.UTC)
	assert.Equal(t, "5. Juni 2026 um 15:04 Uhr", FormatEventDate(d, "de"))
}

func TestFormatEventDateUnsupportedLanguageFallsBackToEnglish(t *testing.T) {
	d := time.Date(2026, time.June, 5, 15, 4, 0, 0, time.UTC)
	assert.Equal(t, "June 5, 2026 at 3:04 PM", FormatEventDate(d, "fr"))
}

func TestInTimezoneConvertsToLocalTime(t *testing.T) {
	// 15:00 UTC is 17:00 in Europe/Berlin during summer (CEST, UTC+2) --
	// notification bodies must show the organizer's entered local time, not
	// the raw stored UTC hour.
	d := time.Date(2026, time.June, 15, 15, 0, 0, 0, time.UTC)
	converted := InTimezone(d, "Europe/Berlin")
	assert.Equal(t, "17:00", converted.Format("15:04"))
	assert.Equal(t, "15. Juni 2026 um 17:00 Uhr", FormatEventDate(converted, "de"))
}

func TestInTimezoneFallsBackToUTCForEmptyOrInvalidZone(t *testing.T) {
	d := time.Date(2026, time.June, 15, 15, 0, 0, 0, time.UTC)
	assert.Equal(t, "15:00", InTimezone(d, "").Format("15:04"))
	assert.Equal(t, "15:00", InTimezone(d, "Not/AZone").Format("15:04"))
}

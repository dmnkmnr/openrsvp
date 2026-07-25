package templates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderRetentionWarning(t *testing.T) {
	html, plain, err := RenderRetentionWarning("Birthday Party", "March 15, 2026", "http://localhost:8080/events")
	require.NoError(t, err)

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
	html, plain, err := RenderRetentionWarning("Garden Party", "April 20, 2026", "")
	require.NoError(t, err)

	assert.Contains(t, html, "Garden Party")
	assert.Contains(t, html, "April 20, 2026")
	assert.NotContains(t, html, "View Event")

	assert.Contains(t, plain, "Garden Party")
	assert.NotContains(t, plain, "visit:")
}

func TestRenderMagicLink(t *testing.T) {
	html, plain, err := RenderMagicLink("http://localhost:8080", "abc123token", 15)
	require.NoError(t, err)

	assert.Contains(t, html, "abc123token")
	assert.Contains(t, html, "http://localhost:8080")
	assert.Contains(t, plain, "15 minutes")
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

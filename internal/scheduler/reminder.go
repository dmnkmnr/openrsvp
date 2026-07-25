package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/yannkr/openrsvp/internal/calendar"
	"github.com/yannkr/openrsvp/internal/database"
	"github.com/yannkr/openrsvp/internal/notification"
	"github.com/yannkr/openrsvp/internal/notification/templates"
	"github.com/yannkr/openrsvp/internal/rsvp"
)

// TemplateResolver resolves the effective subject/body for a guest
// notification message type on an event, in the given language (organizer
// override if set, otherwise the language default). Satisfied by
// *messagetemplate.Service.Resolve.
type TemplateResolver func(ctx context.Context, eventID, messageType, lang string) (subject, body string, err error)

// ReminderJob polls for due reminders and sends notifications to the
// appropriate attendees.
type ReminderJob struct {
	store            *ReminderStore
	db               database.DB
	notifService     *notification.Service
	baseURL          string
	logger           zerolog.Logger
	resolveTemplate  TemplateResolver
}

// NewReminderJob creates a new ReminderJob.
func NewReminderJob(store *ReminderStore, db database.DB, notifService *notification.Service, baseURL string, logger zerolog.Logger) *ReminderJob {
	return &ReminderJob{
		store:        store,
		db:           db,
		notifService: notifService,
		baseURL:      baseURL,
		logger:       logger,
	}
}

// SetTemplateResolver registers the function used to resolve the reminder
// message's subject/body template (organizer override or language default).
// When unset, sendToAttendee falls back to a hardcoded English default.
func (j *ReminderJob) SetTemplateResolver(fn TemplateResolver) {
	j.resolveTemplate = fn
}

// Name returns the job identifier.
func (j *ReminderJob) Name() string {
	return "reminder"
}

// Interval returns how often this job runs.
func (j *ReminderJob) Interval() time.Duration {
	return 30 * time.Second
}

// Run executes one iteration of the reminder job: finds due reminders,
// claims them for processing, sends notifications, and updates status.
func (j *ReminderJob) Run(ctx context.Context) error {
	due, err := j.store.FindDue(ctx)
	if err != nil {
		return fmt.Errorf("find due reminders: %w", err)
	}

	if len(due) == 0 {
		return nil
	}

	j.logger.Info().Int("count", len(due)).Msg("found due reminders")

	for _, reminder := range due {
		if err := j.processReminder(ctx, reminder); err != nil {
			j.logger.Error().
				Err(err).
				Str("reminder_id", reminder.ID).
				Str("event_id", reminder.EventID).
				Msg("failed to process reminder")
		}
	}

	return nil
}

// processReminder claims a single reminder, finds target attendees, sends
// notifications, and updates the reminder status.
func (j *ReminderJob) processReminder(ctx context.Context, reminder *Reminder) error {
	// Claim the reminder to prevent duplicate processing.
	claimed, err := j.store.ClaimForProcessing(ctx, reminder.ID)
	if err != nil {
		return fmt.Errorf("claim reminder: %w", err)
	}
	if !claimed {
		// Another worker already claimed this reminder.
		return nil
	}

	// Find attendees in the target group.
	attendees, err := j.findTargetAttendees(ctx, reminder.EventID, reminder.TargetGroup)
	if err != nil {
		if setErr := j.store.SetStatus(ctx, reminder.ID, "failed"); setErr != nil {
			j.logger.Error().Err(setErr).Str("reminder_id", reminder.ID).Msg("failed to set reminder status to failed")
		}
		return fmt.Errorf("find target attendees: %w", err)
	}

	if len(attendees) == 0 {
		j.logger.Info().
			Str("reminder_id", reminder.ID).
			Str("target_group", reminder.TargetGroup).
			Msg("no attendees in target group, marking as sent")
		return j.store.SetStatus(ctx, reminder.ID, "sent")
	}

	// Look up event details for the email template.
	ev, err := j.lookupEvent(ctx, reminder.EventID)
	if err != nil {
		if setErr := j.store.SetStatus(ctx, reminder.ID, "failed"); setErr != nil {
			j.logger.Error().Err(setErr).Str("reminder_id", reminder.ID).Msg("failed to set reminder status to failed")
		}
		return fmt.Errorf("lookup event for reminder: %w", err)
	}

	// Send notifications to each attendee.
	var sendErrors int
	for _, attendee := range attendees {
		if err := j.sendToAttendee(ctx, reminder, attendee, ev); err != nil {
			sendErrors++
			j.logger.Error().
				Err(err).
				Str("reminder_id", reminder.ID).
				Str("attendee_id", attendee.id).
				Msg("failed to send reminder to attendee")
		}
	}

	// Mark the reminder based on results.
	if sendErrors == len(attendees) {
		return j.store.SetStatus(ctx, reminder.ID, "failed")
	}

	return j.store.SetStatus(ctx, reminder.ID, "sent")
}

// attendeeTarget holds the minimal info needed to send a notification.
type attendeeTarget struct {
	id            string
	name          string
	email         *string
	phone         *string
	rsvpToken     string
	contactMethod string
}

// findTargetAttendees queries for attendees matching the reminder's target
// group. The target_group field filters by RSVP status ('all' means everyone).
func (j *ReminderJob) findTargetAttendees(ctx context.Context, eventID, targetGroup string) ([]attendeeTarget, error) {
	var query string
	var args []any

	if targetGroup == "all" {
		query = `SELECT id, name, email, phone, rsvp_token, contact_method FROM attendees WHERE event_id = ?`
		args = []any{eventID}
	} else {
		query = `SELECT id, name, email, phone, rsvp_token, contact_method FROM attendees WHERE event_id = ? AND rsvp_status = ?`
		args = []any{eventID, targetGroup}
	}

	rows, err := j.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query attendees: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var attendees []attendeeTarget
	for rows.Next() {
		var a attendeeTarget
		var email, phone *string
		if err := rows.Scan(&a.id, &a.name, &email, &phone, &a.rsvpToken, &a.contactMethod); err != nil {
			return nil, fmt.Errorf("scan attendee: %w", err)
		}
		a.email = email
		a.phone = phone
		attendees = append(attendees, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendees: %w", err)
	}

	return attendees, nil
}

// eventInfo holds the minimal event data needed to render reminder emails.
type eventInfo struct {
	id          string
	title       string
	description string
	eventDate   time.Time
	endDate     *time.Time
	location    string
	timezone    string
	shareToken  string
	language    string
}

// lookupEvent fetches event details needed for rendering the reminder template.
func (j *ReminderJob) lookupEvent(ctx context.Context, eventID string) (*eventInfo, error) {
	var info eventInfo
	var description, timezone *string
	var eventDate string
	var endDate *string
	err := j.db.QueryRowContext(ctx,
		`SELECT id, title, description, event_date, end_date, location, timezone, share_token, language FROM events WHERE id = ?`,
		eventID,
	).Scan(&info.id, &info.title, &description, &eventDate, &endDate, &info.location, &timezone, &info.shareToken, &info.language)
	if err != nil {
		return nil, fmt.Errorf("lookup event %s: %w", eventID, err)
	}
	if description != nil {
		info.description = *description
	}
	if timezone != nil {
		info.timezone = *timezone
	}
	// Timestamps are stored as RFC3339 TEXT; parse them like every other
	// scanner in the codebase rather than relying on driver time conversion.
	info.eventDate, err = time.Parse(time.RFC3339, eventDate)
	if err != nil {
		return nil, fmt.Errorf("parse event_date for event %s: %w", eventID, err)
	}
	if endDate != nil && *endDate != "" {
		parsed, perr := time.Parse(time.RFC3339, *endDate)
		if perr != nil {
			return nil, fmt.Errorf("parse end_date for event %s: %w", eventID, perr)
		}
		info.endDate = &parsed
	}
	return &info, nil
}

// sendToAttendee sends a reminder notification to a single attendee via their
// preferred channel. Respects the attendee's chosen contact method (email or
// sms); falls back to whichever channel actually has data when the preferred
// one is unreachable (e.g. an organizer removed the attendee's email after
// they'd chosen "email" as their preference).
func (j *ReminderJob) sendToAttendee(ctx context.Context, reminder *Reminder, attendee attendeeTarget, ev *eventInfo) error {
	lang := ev.language
	if lang == "" {
		lang = "en"
	}

	// The organizer's per-reminder custom message (if set when scheduling this
	// specific reminder) takes precedence over the event's saved/default
	// reminder template; the subject still comes from the language default
	// since it was never organizer-editable per-reminder.
	var subjectTpl, bodyTpl string
	if reminder.Message != "" {
		subjectTpl = templates.DefaultFor(reminderMessageType, lang).Subject
		bodyTpl = reminder.Message
	} else if j.resolveTemplate != nil {
		resolved, resolveErr := j.resolveTemplateOrDefault(ctx, reminder.EventID, lang)
		subjectTpl, bodyTpl = resolved.Subject, resolved.Body
		if resolveErr != nil {
			j.logger.Error().Err(resolveErr).Str("reminder_id", reminder.ID).Msg("failed to resolve reminder template, using language default")
		}
	} else {
		def := templates.DefaultFor(reminderMessageType, lang)
		subjectTpl, bodyTpl = def.Subject, def.Body
	}

	eventDate := ev.eventDate.Format("January 2, 2006 at 3:04 PM")
	location := ev.location
	if location == "" {
		location = "TBD"
	}
	inviteURL := j.baseURL + "/i/" + ev.shareToken
	rsvpLink := inviteURL
	if attendee.rsvpToken != "" {
		rsvpLink = j.baseURL + "/r/" + attendee.rsvpToken
	}

	vars := map[string]string{
		"guestName":  attendee.name,
		"eventTitle": ev.title,
		"eventDate":  eventDate,
		"location":   location,
		"rsvpLink":   rsvpLink,
	}

	hasEmail := attendee.email != nil && *attendee.email != ""
	hasPhone := attendee.phone != nil && *attendee.phone != ""
	wantEmail, wantSMS := rsvp.ResolveChannels(attendee.contactMethod, hasEmail, hasPhone)

	var errs []error

	if wantEmail {
		subject, htmlBody, plainBody, err := templates.RenderNotification(lang, subjectTpl, bodyTpl, vars, rsvpLink, templates.CTALabel(reminderMessageType, lang))
		if err != nil {
			j.logger.Error().Err(err).Str("reminder_id", reminder.ID).Msg("failed to render reminder template, falling back to plain text")
			plain := templates.Interpolate(bodyTpl, vars)
			subject = templates.Interpolate(subjectTpl, vars)
			htmlBody, plainBody = plain, plain
		}

		msg := &notification.Message{
			To:      *attendee.email,
			Subject: subject,
			Body:    htmlBody,
			Plain:   plainBody,
			Lang:    lang,
		}

		// Attach ICS calendar file for attending and maybe attendees,
		// or when the reminder targets all attendees.
		if reminder.TargetGroup == "attending" || reminder.TargetGroup == "maybe" || reminder.TargetGroup == "all" {
			// Use the RSVP management URL when available so the guest can manage
			// their response; fall back to the public invite URL.
			calURL := rsvpLink
			icsData := calendar.GenerateICS(calendar.EventData{
				ID:          ev.id,
				Title:       ev.title,
				Description: ev.description,
				Location:    ev.location,
				EventDate:   ev.eventDate,
				EndDate:     ev.endDate,
				Timezone:    ev.timezone,
				URL:         calURL,
			})
			msg.Attachments = []notification.Attachment{
				{
					Filename:    "event.ics",
					ContentType: "text/calendar; charset=utf-8; method=PUBLISH",
					Data:        []byte(icsData),
				},
			}
		}

		if err := j.notifService.Send(ctx, reminder.EventID, attendee.id, notification.ChannelEmail, msg); err != nil {
			errs = append(errs, err)
		}
	}

	if wantSMS {
		plain := templates.Interpolate(bodyTpl, vars)
		msg := &notification.Message{
			To:   *attendee.phone,
			Body: templates.SMSFrom(plain, 300),
			Lang: lang,
		}
		if err := j.notifService.Send(ctx, reminder.EventID, attendee.id, notification.ChannelSMS, msg); err != nil {
			errs = append(errs, err)
		}
	}

	if !wantEmail && !wantSMS {
		j.logger.Warn().
			Str("attendee_id", attendee.id).
			Msg("attendee has no email or phone for notification")
	}

	return errors.Join(errs...)
}

// reminderMessageType is the message type key used to resolve reminder
// templates. It must match messagetemplate.TypeReminder's value.
const reminderMessageType = "reminder"

// resolvedTemplate mirrors templates.TemplateDefault for use with the
// resolveTemplate callback's plain (subject, body, error) return shape.
type resolvedTemplate struct {
	Subject string
	Body    string
}

// resolveTemplateOrDefault calls the registered TemplateResolver, falling
// back to the language default on error.
func (j *ReminderJob) resolveTemplateOrDefault(ctx context.Context, eventID, lang string) (resolvedTemplate, error) {
	subject, body, err := j.resolveTemplate(ctx, eventID, reminderMessageType, lang)
	if err != nil {
		def := templates.DefaultFor(reminderMessageType, lang)
		return resolvedTemplate{Subject: def.Subject, Body: def.Body}, err
	}
	return resolvedTemplate{Subject: subject, Body: body}, nil
}

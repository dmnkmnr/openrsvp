package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yannkr/openrsvp/internal/auth"
	"github.com/yannkr/openrsvp/internal/calendar"
	"github.com/yannkr/openrsvp/internal/comment"
	"github.com/yannkr/openrsvp/internal/config"
	"github.com/yannkr/openrsvp/internal/database"
	"github.com/yannkr/openrsvp/internal/event"
	"github.com/yannkr/openrsvp/internal/feedback"
	"github.com/yannkr/openrsvp/internal/instanceconfig"
	"github.com/yannkr/openrsvp/internal/invite"
	"github.com/yannkr/openrsvp/internal/message"
	"github.com/yannkr/openrsvp/internal/messagetemplate"
	"github.com/yannkr/openrsvp/internal/notification"
	"github.com/yannkr/openrsvp/internal/notification/templates"
	"github.com/yannkr/openrsvp/internal/question"
	"github.com/yannkr/openrsvp/internal/rsvp"
	"github.com/yannkr/openrsvp/internal/scheduler"
	"github.com/yannkr/openrsvp/internal/security"
	"github.com/yannkr/openrsvp/internal/stats"
	"github.com/yannkr/openrsvp/internal/suppression"
	"github.com/yannkr/openrsvp/internal/webhook"
)

// Server is the main HTTP server for OpenRSVP.
type Server struct {
	cfg                   *config.Config
	db                    database.DB
	logger                zerolog.Logger
	http                  *http.Server
	authHandler           *auth.Handler
	eventHandler          *event.Handler
	seriesHandler         *event.SeriesHandler
	rsvpHandler           *rsvp.Handler
	inviteHandler         *invite.Handler
	messageHandler        *message.Handler
	messageTemplateHandler *messagetemplate.Handler
	questionHandler       *question.Handler
	feedbackHandler       *feedback.Handler
	reminderHandler       *scheduler.Handler
	commentHandler        *comment.Handler
	webhookHandler        *webhook.Handler
	notifHandler          *notification.Handler
	notifService          *notification.Service
	statsHandler          *stats.Handler
	suppressionHandler    *suppression.Handler
	instanceConfigHandler *instanceconfig.Handler
	scheduler             *scheduler.Scheduler
	securityMw            *security.Middleware
	uploadsDir            string
	smsEnabled            bool
}

// commentEventStoreAdapter adapts event.Service to comment.EventStore.
type commentEventStoreAdapter struct {
	svc *event.Service
}

func (a *commentEventStoreAdapter) FindByShareToken(ctx context.Context, shareToken string) (*comment.Event, error) {
	ev, err := a.svc.GetByShareToken(ctx, shareToken)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, nil
	}
	return &comment.Event{
		ID:              ev.ID,
		Status:          ev.Status,
		CommentsEnabled: ev.CommentsEnabled,
	}, nil
}

// commentRSVPStoreAdapter adapts rsvp.Store to comment.RSVPStore.
type commentRSVPStoreAdapter struct {
	store *rsvp.Store
}

func (a *commentRSVPStoreAdapter) FindByToken(ctx context.Context, token string) (*comment.Attendee, error) {
	att, err := a.store.FindByRSVPToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, nil
	}
	return &comment.Attendee{
		ID:      att.ID,
		EventID: att.EventID,
		Name:    att.Name,
	}, nil
}

// New creates a new Server instance.
func New(cfg *config.Config, db database.DB, logger zerolog.Logger) *Server {
	// Wire up auth layer.
	authStore := auth.NewStore(db)
	authService := auth.NewService(authStore, cfg, logger)
	authHandler := auth.NewHandler(authService, cfg, logger)
	authMiddleware := auth.RequireAuth(authService)

	organizerFromCtx := func(ctx context.Context) (string, bool) {
		org := auth.OrganizerFromContext(ctx)
		if org == nil {
			return "", false
		}
		return org.ID, true
	}

	// Build the notification provider registry early: SMS availability must
	// reflect whether an SMS provider actually registered (valid credentials),
	// not just that NOTIFICATION_SMS_PROVIDER names one. Checking only the env
	// var's presence would let a misconfigured provider (e.g. a missing Twilio
	// credential) pass as "enabled", so organizers could pick "phone only"
	// contact requirements and show guests a phone field that no notification
	// can ever actually reach.
	notifRegistry := buildNotificationRegistry(cfg, logger)
	smsEnabled := notifRegistry.Has(notification.ChannelSMS)

	// Wire up event layer.
	eventStore := event.NewStore(db)
	eventService := event.NewService(eventStore, cfg.DefaultRetentionDays)

	// Wire up co-host store and set it on the event service.
	cohostStore := event.NewCoHostStore(db)
	eventService.SetCoHostStore(cohostStore)

	// Organizer lookup by email for co-host management.
	organizerLookupByEmail := event.OrganizerLookupByEmail(func(ctx context.Context, email string) (string, string, error) {
		org, err := authStore.FindOrganizerByEmail(ctx, email)
		if err != nil {
			return "", "", err
		}
		if org == nil {
			return "", "", nil
		}
		return org.ID, org.Name, nil
	})

	eventHandler := event.NewHandler(
		eventService, authMiddleware, event.OrganizerFromCtx(organizerFromCtx), logger,
		event.WithCoHostStore(cohostStore),
		event.WithOrganizerLookup(organizerLookupByEmail),
		event.WithMaxCoHosts(cfg.MaxCoHostsPerEvent),
	)

	// Wire up event series layer.
	seriesStore := event.NewSeriesStore(db)
	seriesService := event.NewSeriesService(seriesStore, eventStore, eventService, cfg.DefaultRetentionDays, logger)
	seriesHandler := event.NewSeriesHandler(seriesService, authMiddleware, event.OrganizerFromCtx(organizerFromCtx), logger)

	// checkEventOwner verifies that the given organizer can manage the event
	// (either as owner or co-host).
	// Returns nil if the organizer can manage the event; a non-nil error otherwise.
	checkEventOwner := func(ctx context.Context, eventID, organizerID string) error {
		canManage, err := eventService.CanManageEvent(ctx, eventID, organizerID)
		if err != nil {
			return err
		}
		if !canManage {
			return fmt.Errorf("event not found")
		}
		return nil
	}

	// Wire up per-event message template overrides (customizable guest
	// notification subject/body, per event language). Created early since
	// several guest-notification callbacks below need messageTemplateService.
	messageTemplateStore := messagetemplate.NewStore(db)
	messageTemplateService := messagetemplate.NewService(messageTemplateStore, logger)
	// resolveTemplateOrDefault wraps messageTemplateService.Resolve, falling
	// back to the language default on error so a transient DB error never
	// results in an email/SMS being sent with an empty subject/body.
	resolveTemplateOrDefault := func(ctx context.Context, eventID, messageType, lang string) (subject, body string) {
		subject, body, err := messageTemplateService.Resolve(ctx, eventID, messageType, lang)
		if err != nil {
			logger.Error().Err(err).Str("event_id", eventID).Str("message_type", messageType).Msg("failed to resolve message template, using language default")
			def := templates.DefaultFor(messageType, lang)
			return def.Subject, def.Body
		}
		return subject, body
	}
	checkEventOwnerWithLang := func(ctx context.Context, eventID, organizerID string) (string, error) {
		if err := checkEventOwner(ctx, eventID, organizerID); err != nil {
			return "", err
		}
		ev, err := eventStore.FindByID(ctx, eventID)
		if err != nil {
			return "", err
		}
		if ev == nil {
			return "", fmt.Errorf("event not found")
		}
		return ev.Language, nil
	}
	messageTemplateHandler := messagetemplate.NewHandler(messageTemplateService, authMiddleware, messagetemplate.OrganizerFromCtx(organizerFromCtx), messagetemplate.EventOwnershipChecker(checkEventOwnerWithLang), logger)

	// Ensure uploads directory exists.
	uploadsDir := cfg.UploadsDir
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		logger.Error().Err(err).Str("dir", uploadsDir).Msg("failed to create uploads directory")
	}

	// Wire up invite layer (before RSVP since RSVP depends on it).
	inviteStore := invite.NewStore(db)
	inviteService := invite.NewService(inviteStore, uploadsDir)
	inviteHandler := invite.NewHandler(inviteService, authMiddleware, invite.OrganizerFromCtx(organizerFromCtx), uploadsDir, invite.EventOwnershipChecker(checkEventOwner), logger)

	// Configure SMS availability on event service.
	eventService.SetSMSEnabled(smsEnabled)

	// Wire up RSVP layer.
	rsvpStore := rsvp.NewStore(db)
	rsvpService := rsvp.NewService(rsvpStore, eventService, inviteService, logger)
	rsvpService.SetSMSEnabled(smsEnabled)
	rsvpService.SetBaseURL(cfg.BaseURL)
	rsvpHandler := rsvp.NewHandler(rsvpService, authMiddleware, rsvp.OrganizerFromCtx(organizerFromCtx), rsvp.EventOwnershipChecker(checkEventOwner), logger)

	// Wire up question layer.
	questionStore := question.NewStore(db)
	questionService := question.NewService(questionStore)
	questionHandler := question.NewHandler(questionService, authMiddleware, question.OrganizerFromCtx(organizerFromCtx), question.EventOwnershipChecker(checkEventOwner), logger)

	// Wire question validation and listing into the RSVP service.
	rsvpService.SetValidateAnswers(questionService.ValidateAndSaveAnswers)
	rsvpService.SetListQuestions(func(ctx context.Context, eventID string) (any, error) {
		return questionService.ListByEvent(ctx, eventID)
	})
	rsvpService.SetGetAnswers(func(ctx context.Context, attendeeID string) (any, error) {
		return questionService.GetAnswersForAttendee(ctx, attendeeID)
	})
	rsvpService.SetGetExportQuestions(func(ctx context.Context, eventID string) (*rsvp.ExportQuestionsData, error) {
		questions, err := questionService.ListByEvent(ctx, eventID)
		if err != nil {
			return nil, err
		}
		if len(questions) == 0 {
			return nil, nil
		}
		answersByEvent, err := questionService.GetAnswersByEvent(ctx, eventID)
		if err != nil {
			return nil, err
		}
		data := &rsvp.ExportQuestionsData{
			Labels:            make([]string, len(questions)),
			QuestionIDs:       make([]string, len(questions)),
			AnswersByAttendee: make(map[string]map[string]string),
		}
		for i, q := range questions {
			data.Labels[i] = q.Label
			data.QuestionIDs[i] = q.ID
		}
		for attendeeID, answers := range answersByEvent {
			if data.AnswersByAttendee[attendeeID] == nil {
				data.AnswersByAttendee[attendeeID] = make(map[string]string)
			}
			for _, a := range answers {
				data.AnswersByAttendee[attendeeID][a.QuestionID] = a.Answer
			}
		}
		return data, nil
	})

	// Wire up comment/guestbook layer.
	commentStore := comment.NewStore(db)
	commentEventAdapter := &commentEventStoreAdapter{svc: eventService}
	commentRSVPAdapter := &commentRSVPStoreAdapter{store: rsvpStore}
	commentService := comment.NewService(commentStore, commentEventAdapter, commentRSVPAdapter)
	commentHandler := comment.NewHandler(commentService, authMiddleware, comment.OrganizerFromCtx(organizerFromCtx), comment.EventOwnershipChecker(checkEventOwner), logger)

	// Wire up webhook layer.
	webhookStore := webhook.NewStore(db)
	webhookService := webhook.NewService(webhookStore, logger, !cfg.IsDevelopment())
	webhookDispatcher := webhook.NewDispatcher(webhookStore, logger)
	webhookHandler := webhook.NewHandler(webhookService, webhookDispatcher, authMiddleware, webhook.OrganizerFromCtx(organizerFromCtx), webhook.EventOwnershipChecker(checkEventOwner), logger)

	// Wire up email suppression / unsubscribe layer (consumed by notification).
	suppressionStore := suppression.NewStore(db)
	suppressionService := suppression.NewService(suppressionStore)
	suppressionHandler := suppression.NewHandler(suppressionService, logger)

	// Wire up notification layer. notifRegistry was built earlier (see above)
	// so SMS availability could be computed before eventService/rsvpService
	// construction.
	notifService := notification.NewServiceWithOptions(notifRegistry, db, logger, notification.Options{
		BaseURL:             cfg.BaseURL,
		OpenTrackingEnabled: cfg.EmailOpenTrackingEnabled,
		Suppression:         suppressionService,
	})

	// Wire up notification tracking layer.
	trackingService := notification.NewTrackingService(db, logger)
	notifHandler := notification.NewHandler(trackingService, notifService, suppressionService, authMiddleware, notification.OrganizerFromCtx(organizerFromCtx), notification.EventOwnershipChecker(checkEventOwner), logger)

	// Wire email sending into auth service (breaks circular dep via function).
	if notifRegistry.Has(notification.ChannelEmail) {
		authService.SetEmailSender(func(ctx context.Context, to, subject, htmlBody, plainBody string) error {
			provider, err := notifRegistry.Get(notification.ChannelEmail)
			if err != nil {
				return err
			}
			_, sendErr := provider.Send(ctx, &notification.Message{
				To:      to,
				Subject: subject,
				Body:    htmlBody,
				Plain:   plainBody,
			})
			return sendErr
		})
	}

	// Wire email sending into RSVP service (for RSVP lookup magic links).
	if notifRegistry.Has(notification.ChannelEmail) {
		rsvpService.SetEmailSender(func(ctx context.Context, to, subject, htmlBody, plainBody string) error {
			provider, err := notifRegistry.Get(notification.ChannelEmail)
			if err != nil {
				return err
			}
			_, sendErr := provider.Send(ctx, &notification.Message{
				To:      to,
				Subject: subject,
				Body:    htmlBody,
				Plain:   plainBody,
			})
			return sendErr
		})
	}

	// Wire SMS sending into RSVP service (for RSVP lookup magic links, so
	// phone-only guests can find their RSVP too).
	if notifRegistry.Has(notification.ChannelSMS) {
		rsvpService.SetSMSSender(func(ctx context.Context, to, body string) error {
			provider, err := notifRegistry.Get(notification.ChannelSMS)
			if err != nil {
				return err
			}
			_, sendErr := provider.Send(ctx, &notification.Message{
				To:   to,
				Body: body,
			})
			return sendErr
		})
	}

	// Wire RSVP confirmation notifications into the RSVP service.
	if notifRegistry.Has(notification.ChannelEmail) || notifRegistry.Has(notification.ChannelSMS) {
		rsvpService.SetNotifyRSVP(func(ctx context.Context, eventID string, attendee *rsvp.Attendee) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("rsvp notify: failed to get event")
				return
			}

			// Dispatch webhook for RSVP event.
			go webhookDispatcher.Dispatch(context.Background(), eventID, "rsvp.created", map[string]any{
				"attendeeId":   attendee.ID,
				"attendeeName": attendee.Name,
				"rsvpStatus":   attendee.RSVPStatus,
				"eventId":      eventID,
			})

			eventDate := templates.FormatEventDate(templates.InTimezone(ev.EventDate, ev.Timezone), ev.Language)
			location := ev.Location
			if location == "" {
				location = "TBD"
			}

			// Send confirmation to the attendee via their preferred channel,
			// falling back to whichever channel actually has data.
			hasEmail := attendee.Email != nil && *attendee.Email != ""
			hasPhone := attendee.Phone != nil && *attendee.Phone != ""
			modifyURL := cfg.BaseURL + "/r/" + attendee.RSVPToken
			vars := map[string]string{
				"guestName":  attendee.Name,
				"eventTitle": ev.Title,
				"eventDate":  eventDate,
				"location":   location,
				"rsvpStatus": templates.DisplayStatusLocalized(ev.Language, attendee.RSVPStatus),
				"rsvpLink":   modifyURL,
			}
			subjectTpl, bodyTpl := resolveTemplateOrDefault(ctx, eventID, messagetemplate.TypeRSVPConfirmation, ev.Language)
			wantEmail, wantSMS := rsvp.ResolveChannels(attendee.ContactMethod, hasEmail, hasPhone)
			if wantEmail {
				subject, htmlBody, plainBody, err := templates.RenderNotification(ev.Language, subjectTpl, bodyTpl, vars, modifyURL, templates.CTALabel(messagetemplate.TypeRSVPConfirmation, ev.Language))
				if err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("rsvp notify: failed to render attendee template")
				} else {
					confirmMsg := &notification.Message{
						To:      *attendee.Email,
						Subject: subject,
						Body:    htmlBody,
						Plain:   plainBody,
						Lang:    ev.Language,
					}

					// Attach ICS calendar file for attending and maybe RSVPs.
					// Use the RSVP management URL so the guest can manage their response.
					if attendee.RSVPStatus == "attending" || attendee.RSVPStatus == "maybe" {
						rsvpURL := cfg.BaseURL + "/r/" + attendee.RSVPToken
						icsData := calendar.GenerateICS(calendar.EventData{
							ID:          ev.ID,
							Title:       ev.Title,
							Description: ev.Description,
							Location:    ev.Location,
							EventDate:   ev.EventDate,
							EndDate:     ev.EndDate,
							Timezone:    ev.Timezone,
							URL:         rsvpURL,
						})
						confirmMsg.Attachments = []notification.Attachment{
							{
								Filename:    "event.ics",
								ContentType: "text/calendar; charset=utf-8; method=PUBLISH",
								Data:        []byte(icsData),
							},
						}
					}

					if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelEmail, confirmMsg); err != nil {
						logger.Error().Err(err).Str("attendee_email", *attendee.Email).Msg("rsvp notify: failed to send attendee email")
					}
				}
			}
			if wantSMS {
				smsBody := templates.SMSFrom(templates.Interpolate(bodyTpl, vars)+"\n\n"+modifyURL, 300)
				if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelSMS, &notification.Message{
					To:   *attendee.Phone,
					Body: smsBody,
					Lang: ev.Language,
				}); err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("rsvp notify: failed to send attendee SMS")
				}
			}

			// Notify the organizer about the new RSVP. Organizers always
			// authenticate via email, so this stays email-only.
			if !notifRegistry.Has(notification.ChannelEmail) {
				return
			}
			organizer, err := authStore.FindOrganizerByID(ctx, ev.OrganizerID)
			if err != nil {
				logger.Error().Err(err).Str("organizer_id", ev.OrganizerID).Msg("rsvp notify: failed to get organizer")
				return
			}
			if organizer == nil || organizer.Email == "" {
				return
			}

			guestEmail := ""
			if attendee.Email != nil {
				guestEmail = *attendee.Email
			}
			guestPhone := ""
			if attendee.Phone != nil {
				guestPhone = *attendee.Phone
			}
			dashboardURL := cfg.BaseURL + "/events/" + eventID

			subjectPrefix, htmlBody, plainBody, err := templates.RenderOrganizerRSVPNotification(
				organizer.Language, ev.Title, attendee.Name, attendee.RSVPStatus,
				guestEmail, guestPhone, attendee.PlusOnes, attendee.PlusOnesChildren, dashboardURL,
			)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("rsvp notify: failed to render organizer template")
				return
			}

			if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelEmail, &notification.Message{
				To:      organizer.Email,
				Subject: subjectPrefix + " — " + attendee.Name + " — " + ev.Title + " (" + templates.InTimezone(ev.EventDate, ev.Timezone).Format("Jan 2") + ")",
				Body:    htmlBody,
				Plain:   plainBody,
				Lang:    organizer.Language,
			}); err != nil {
				logger.Error().Err(err).Str("organizer_email", organizer.Email).Msg("rsvp notify: failed to send organizer email")
			}
		})
	}

	// Wire import invitation notifications into the RSVP service.
	if notifRegistry.Has(notification.ChannelEmail) || notifRegistry.Has(notification.ChannelSMS) {
		rsvpService.SetOnImportInvite(func(ctx context.Context, eventID string, attendee *rsvp.Attendee) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("import invite: failed to get event")
				return
			}

			eventDate := templates.FormatEventDate(templates.InTimezone(ev.EventDate, ev.Timezone), ev.Language)
			location := ev.Location
			if location == "" {
				location = "TBD"
			}
			inviteURL := cfg.BaseURL + "/i/" + ev.ShareToken

			vars := map[string]string{
				"guestName":  attendee.Name,
				"eventTitle": ev.Title,
				"eventDate":  eventDate,
				"location":   location,
				"rsvpLink":   inviteURL,
			}
			subjectTpl, bodyTpl := resolveTemplateOrDefault(ctx, eventID, messagetemplate.TypeImportInvite, ev.Language)

			// Respect the attendee's contact method preference (CSV import
			// infers it from what the row provided), falling back to whichever
			// channel actually has data.
			hasEmail := attendee.Email != nil && *attendee.Email != ""
			hasPhone := attendee.Phone != nil && *attendee.Phone != ""
			wantEmail, wantSMS := rsvp.ResolveChannels(attendee.ContactMethod, hasEmail, hasPhone)
			if wantEmail {
				subject, htmlBody, plainBody, err := templates.RenderNotification(ev.Language, subjectTpl, bodyTpl, vars, inviteURL, templates.CTALabel(messagetemplate.TypeImportInvite, ev.Language))
				if err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("import invite: failed to render template")
				} else if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelEmail, &notification.Message{
					To:      *attendee.Email,
					Subject: subject,
					Body:    htmlBody,
					Plain:   plainBody,
					Lang:    ev.Language,
				}); err != nil {
					logger.Error().Err(err).Str("attendee_email", *attendee.Email).Msg("import invite: failed to send email")
				}
			}
			if wantSMS {
				smsBody := templates.SMSFrom(templates.Interpolate(bodyTpl, vars)+"\n\n"+inviteURL, 300)
				if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelSMS, &notification.Message{
					To:   *attendee.Phone,
					Body: smsBody,
					Lang: ev.Language,
				}); err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("import invite: failed to send SMS")
				}
			}
		})
	}

	// Wire co-host invitation notifications into the event handler.
	if notifRegistry.Has(notification.ChannelEmail) {
		eventHandler.SetNotifyCoHost(func(ctx context.Context, coHostEmail, eventID, addedByOrganizerID string) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("cohost notify: failed to get event")
				return
			}

			organizer, err := authStore.FindOrganizerByID(ctx, addedByOrganizerID)
			if err != nil || organizer == nil {
				logger.Error().Err(err).Str("organizer_id", addedByOrganizerID).Msg("cohost notify: failed to get organizer")
				return
			}

			eventDate := templates.InTimezone(ev.EventDate, ev.Timezone).Format("January 2, 2006 at 3:04 PM")
			location := ev.Location
			if location == "" {
				location = "TBD"
			}
			dashboardURL := cfg.BaseURL + "/events/" + eventID

			// The email's recipient is the invited co-host, not the inviting
			// organizer -- look up their own saved language preference if
			// they're already a registered organizer; default to English if
			// they aren't (first-time invitees have no account yet).
			recipientLang := "en"
			if recipient, err := authStore.FindOrganizerByEmail(ctx, coHostEmail); err == nil && recipient != nil {
				recipientLang = recipient.Language
			}

			subject, htmlBody, plainBody, err := templates.RenderCoHostInvitation(recipientLang, ev.Title, eventDate, location, organizer.Name, dashboardURL)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("cohost notify: failed to render template")
				return
			}

			if err := notifService.Send(ctx, eventID, addedByOrganizerID, notification.ChannelEmail, &notification.Message{
				To:      coHostEmail,
				Subject: subject + " — " + ev.Title,
				Body:    htmlBody,
				Plain:   plainBody,
				Lang:    recipientLang,
			}); err != nil {
				logger.Error().Err(err).Str("cohost_email", coHostEmail).Msg("cohost notify: failed to send email")
			}
		})
	}

	// Wire waitlist promotion notifications into the RSVP service.
	if notifRegistry.Has(notification.ChannelEmail) || notifRegistry.Has(notification.ChannelSMS) {
		rsvpService.SetNotifyWaitlistPromotion(func(ctx context.Context, eventID string, attendee *rsvp.Attendee) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("waitlist promote: failed to get event")
				return
			}

			eventDate := templates.FormatEventDate(templates.InTimezone(ev.EventDate, ev.Timezone), ev.Language)
			location := ev.Location
			if location == "" {
				location = "TBD"
			}
			modifyURL := cfg.BaseURL + "/r/" + attendee.RSVPToken

			vars := map[string]string{
				"guestName":  attendee.Name,
				"eventTitle": ev.Title,
				"eventDate":  eventDate,
				"location":   location,
				"rsvpLink":   modifyURL,
			}
			subjectTpl, bodyTpl := resolveTemplateOrDefault(ctx, eventID, messagetemplate.TypeWaitlistPromotion, ev.Language)

			// Respect the attendee's contact method preference, falling back
			// to whichever channel actually has data.
			hasEmail := attendee.Email != nil && *attendee.Email != ""
			hasPhone := attendee.Phone != nil && *attendee.Phone != ""
			wantEmail, wantSMS := rsvp.ResolveChannels(attendee.ContactMethod, hasEmail, hasPhone)
			if wantEmail {
				subject, htmlBody, plainBody, err := templates.RenderNotification(ev.Language, subjectTpl, bodyTpl, vars, modifyURL, templates.CTALabel(messagetemplate.TypeWaitlistPromotion, ev.Language))
				if err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("waitlist promote: failed to render template")
				} else if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelEmail, &notification.Message{
					To:      *attendee.Email,
					Subject: subject,
					Body:    htmlBody,
					Plain:   plainBody,
					Lang:    ev.Language,
				}); err != nil {
					logger.Error().Err(err).Str("attendee_email", *attendee.Email).Msg("waitlist promote: failed to send email")
				}
			}
			if wantSMS {
				smsBody := templates.SMSFrom(templates.Interpolate(bodyTpl, vars)+"\n\n"+modifyURL, 300)
				if err := notifService.Send(ctx, eventID, attendee.ID, notification.ChannelSMS, &notification.Message{
					To:   *attendee.Phone,
					Body: smsBody,
					Lang: ev.Language,
				}); err != nil {
					logger.Error().Err(err).Str("attendee_id", attendee.ID).Msg("waitlist promote: failed to send SMS")
				}
			}
		})
	}

	// Wire up feedback layer.
	feedbackSvc := feedback.NewService(cfg.FeedbackGitHubToken, cfg.FeedbackGitHubRepo, cfg.FeedbackEmail)
	if cfg.FeedbackGitHubToken == "" && cfg.FeedbackEmail == "" {
		logger.Warn().Msg("feedback: no channel configured (set FEEDBACK_GITHUB_TOKEN+FEEDBACK_GITHUB_REPO or FEEDBACK_EMAIL) — submissions will be silently discarded")
	}
	if notifRegistry.Has(notification.ChannelEmail) {
		feedbackSvc.SetEmailSender(func(ctx context.Context, to, subject, body, plain string) error {
			provider, err := notifRegistry.Get(notification.ChannelEmail)
			if err != nil {
				return err
			}
			_, sendErr := provider.Send(ctx, &notification.Message{
				To:      to,
				Subject: subject,
				Body:    body,
				Plain:   plain,
			})
			return sendErr
		})
	}
	organizerEmailFromCtx := func(ctx context.Context) (string, string, bool) {
		org := auth.OrganizerFromContext(ctx)
		if org == nil {
			return "", "", false
		}
		return org.Email, org.Language, true
	}
	feedbackHandler := feedback.NewHandler(feedbackSvc, authMiddleware, feedback.OrganizerFromCtx(organizerEmailFromCtx), logger)

	// Wire up security middleware (created early so rate limiters are available
	// for handler constructors that need them).
	secMw := security.NewMiddleware(security.SecurityConfig{
		AuthRateLimit:    10,
		RSVPRateLimit:    30,
		GeneralRateLimit: 200,
		RateWindow:       1 * time.Minute,
		CSRFExcludePaths: []string{
			"/api/v1/rsvp/public/",
			"/api/v1/auth/magic-link",
			"/api/v1/auth/verify",
			"/api/v1/comments/public/",
			"/api/v1/feedback/public",         // unauthenticated guest bug reports
			"/api/v1/unsubscribe",             // token-based email opt-out (no session)
			"/api/v1/notifications/webhooks/", // inbound SendGrid/SES delivery events
		},
		IsProduction: cfg.Env == "production",
	})

	// Wire up message layer.
	messageStore := message.NewStore(db)
	messageService := message.NewService(messageStore, logger)
	attendeeFromToken := func(ctx context.Context, rsvpToken string) (*message.AttendeeInfo, error) {
		attendee, err := rsvpService.GetByToken(ctx, rsvpToken)
		if err != nil {
			return nil, err
		}
		return &message.AttendeeInfo{ID: attendee.ID, EventID: attendee.EventID}, nil
	}
	messageHandler := message.NewHandler(messageService, authMiddleware, security.RateLimitMiddleware(secMw.RSVPRateLimiter), message.OrganizerFromCtx(organizerFromCtx), attendeeFromToken, message.EventOwnershipChecker(checkEventOwner), logger)

	// Wire up per-event message template overrides (customizable guest
	// notification subject/body, per event language).
	// Wire dispatch into message service so organizer broadcast messages are
	// delivered to attendees via email or SMS.
	if notifRegistry.Has(notification.ChannelEmail) || notifRegistry.Has(notification.ChannelSMS) {
		messageService.SetNotifyAttendees(func(ctx context.Context, eventID, recipientGroup, subject, body string) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("message notify: failed to get event")
				return
			}

			attendees, err := rsvpService.ListByEvent(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("message notify: failed to list attendees")
				return
			}

			inviteURL := cfg.BaseURL + "/i/" + ev.ShareToken
			eventDate := templates.FormatEventDate(templates.InTimezone(ev.EventDate, ev.Timezone), ev.Language)
			location := ev.Location
			if location == "" {
				location = "TBD"
			}

			sent := 0
			for _, a := range attendees {
				// Filter by group.
				if recipientGroup != "all" && a.RSVPStatus != recipientGroup {
					continue
				}

				// Respect each attendee's contact method preference, falling
				// back to whichever channel actually has data.
				hasEmail := a.Email != nil && *a.Email != ""
				hasPhone := a.Phone != nil && *a.Phone != ""
				wantEmail, wantSMS := rsvp.ResolveChannels(a.ContactMethod, hasEmail, hasPhone)
				if wantEmail {
					htmlBody, plainBody, err := templates.RenderEventReminder(ev.Title, eventDate, location, body, inviteURL)
					if err != nil {
						logger.Error().Err(err).Str("attendee_id", a.ID).Msg("message notify: failed to render template")
					} else if err := notifService.Send(ctx, eventID, a.ID, notification.ChannelEmail, &notification.Message{
						To:      *a.Email,
						Subject: subject,
						Body:    htmlBody,
						Plain:   plainBody,
					}); err != nil {
						logger.Error().Err(err).Str("attendee_email", *a.Email).Msg("message notify: failed to send email")
					} else {
						sent++
					}
				}
				if wantSMS {
					if err := notifService.Send(ctx, eventID, a.ID, notification.ChannelSMS, &notification.Message{
						To:   *a.Phone,
						Body: subject + ": " + body,
					}); err != nil {
						logger.Error().Err(err).Str("attendee_id", a.ID).Msg("message notify: failed to send SMS")
					} else {
						sent++
					}
				}
			}

			logger.Info().
				Str("event_id", eventID).
				Str("group", recipientGroup).
				Int("sent", sent).
				Msg("message notify: notifications dispatched")
		})
	}

	// Attendee replies always go to the organizer's email -- organizers
	// always authenticate via email, so this stays email-only.
	if notifRegistry.Has(notification.ChannelEmail) {
		messageService.SetNotifyOrganizer(func(ctx context.Context, eventID, attendeeID, subject, body string) {
			ev, err := eventService.GetByID(ctx, eventID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("attendee notify: failed to get event")
				return
			}

			organizer, err := authStore.FindOrganizerByID(ctx, ev.OrganizerID)
			if err != nil {
				logger.Error().Err(err).Str("organizer_id", ev.OrganizerID).Msg("attendee notify: failed to get organizer")
				return
			}
			if organizer == nil || organizer.Email == "" {
				return
			}

			// Look up attendee name for a personalized notification.
			senderName := "A guest"
			if attendee, err := rsvpStore.FindByID(ctx, attendeeID); err == nil && attendee != nil {
				senderName = attendee.Name
			}

			dashboardURL := cfg.BaseURL + "/events/" + eventID + "/messages"
			emailSubject, htmlBody, plainBody, err := templates.RenderGuestMessageNotification(
				organizer.Language, ev.Title, senderName, subject, body, dashboardURL,
			)
			if err != nil {
				logger.Error().Err(err).Str("event_id", eventID).Msg("attendee notify: failed to render template")
				return
			}

			if err := notifService.Send(ctx, eventID, attendeeID, notification.ChannelEmail, &notification.Message{
				To:      organizer.Email,
				Subject: emailSubject,
				Body:    htmlBody,
				Plain:   plainBody,
				Lang:    organizer.Language,
			}); err != nil {
				logger.Error().Err(err).Str("organizer_email", organizer.Email).Msg("attendee notify: failed to send email")
			}
		})
	}

	// Wire up scheduler and reminder layer.
	reminderStore := scheduler.NewReminderStore(db)
	reminderHandler := scheduler.NewHandler(reminderStore, authMiddleware, scheduler.OrganizerFromCtx(organizerFromCtx), scheduler.EventOwnershipChecker(checkEventOwner), logger)

	// Copy invite card design when an event is duplicated.
	eventService.SetOnDuplicate(func(ctx context.Context, srcEventID, newEventID string) {
		card, err := inviteService.GetByEventID(ctx, srcEventID)
		if err != nil || card == nil {
			return
		}
		_, err = inviteService.Save(ctx, newEventID, invite.SaveInviteRequest{
			TemplateID:     card.TemplateID,
			Heading:        card.Heading,
			Body:           card.Body,
			Footer:         card.Footer,
			PrimaryColor:   card.PrimaryColor,
			SecondaryColor: card.SecondaryColor,
			Font:           card.Font,
			CustomData:     card.CustomData,
		})
		if err != nil {
			logger.Error().Err(err).
				Str("src_event_id", srcEventID).
				Str("new_event_id", newEventID).
				Msg("failed to copy invite card during duplication")
		}
	})

	// Create default reminders (1 week and 3 days before) when an event is published.
	eventService.SetOnPublish(func(ctx context.Context, e *event.Event) {
		go webhookDispatcher.Dispatch(context.Background(), e.ID, "event.published", map[string]any{
			"eventId": e.ID,
			"title":   e.Title,
		})

		// Message is left empty so sendToAttendee resolves the organizer's
		// customized (or language-default) reminder template instead of a
		// fixed English string -- see internal/scheduler/reminder.go.
		offsets := []time.Duration{7 * 24 * time.Hour, 3 * 24 * time.Hour}

		now := time.Now().UTC()
		for _, offset := range offsets {
			remindAt := e.EventDate.Add(-offset)
			if remindAt.Before(now) {
				logger.Debug().
					Str("event_id", e.ID).
					Time("remind_at", remindAt).
					Msg("skipping default reminder (already in the past)")
				continue
			}

			r := &scheduler.Reminder{
				ID:          uuid.Must(uuid.NewV7()).String(),
				EventID:     e.ID,
				RemindAt:    remindAt,
				TargetGroup: "all",
				Status:      "scheduled",
			}
			if err := reminderStore.Create(ctx, r); err != nil {
				logger.Error().Err(err).
					Str("event_id", e.ID).
					Time("remind_at", remindAt).
					Msg("failed to create default reminder")
				continue
			}
			logger.Info().
				Str("event_id", e.ID).
				Str("reminder_id", r.ID).
				Time("remind_at", remindAt).
				Msg("created default reminder")
		}
	})

	// Send cancellation notifications to attending/maybe attendees when an event is cancelled.
	if notifRegistry.Has(notification.ChannelEmail) || notifRegistry.Has(notification.ChannelSMS) {
		eventService.SetOnCancel(func(ctx context.Context, e *event.Event) {
			go webhookDispatcher.Dispatch(context.Background(), e.ID, "event.cancelled", map[string]any{
				"eventId": e.ID,
				"title":   e.Title,
			})

			attendees, err := rsvpService.ListByEvent(ctx, e.ID)
			if err != nil {
				logger.Error().Err(err).Str("event_id", e.ID).Msg("cancel notify: failed to list attendees")
				return
			}

			eventDate := templates.FormatEventDate(templates.InTimezone(e.EventDate, e.Timezone), e.Language)
			location := e.Location
			if location == "" {
				location = "TBD"
			}
			inviteURL := cfg.BaseURL + "/i/" + e.ShareToken
			subjectTpl, bodyTpl := resolveTemplateOrDefault(ctx, e.ID, messagetemplate.TypeCancellation, e.Language)

			sent := 0
			for _, a := range attendees {
				if a.RSVPStatus != "attending" && a.RSVPStatus != "maybe" {
					continue
				}

				vars := map[string]string{
					"guestName":  a.Name,
					"eventTitle": e.Title,
					"eventDate":  eventDate,
					"location":   location,
				}

				// Respect each attendee's contact method preference, falling
				// back to whichever channel actually has data.
				hasEmail := a.Email != nil && *a.Email != ""
				hasPhone := a.Phone != nil && *a.Phone != ""
				wantEmail, wantSMS := rsvp.ResolveChannels(a.ContactMethod, hasEmail, hasPhone)
				if wantEmail {
					subject, htmlBody, plainBody, err := templates.RenderNotification(e.Language, subjectTpl, bodyTpl, vars, inviteURL, templates.CTALabel(messagetemplate.TypeCancellation, e.Language))
					if err != nil {
						logger.Error().Err(err).Str("attendee_id", a.ID).Msg("cancel notify: failed to render template")
					} else if err := notifService.Send(ctx, e.ID, a.ID, notification.ChannelEmail, &notification.Message{
						To:      *a.Email,
						Subject: subject,
						Body:    htmlBody,
						Plain:   plainBody,
						Lang:    e.Language,
					}); err != nil {
						logger.Error().Err(err).Str("attendee_email", *a.Email).Msg("cancel notify: failed to send email")
					} else {
						sent++
					}
				}
				if wantSMS {
					smsBody := templates.SMSFrom(templates.Interpolate(bodyTpl, vars)+"\n\n"+inviteURL, 300)
					if err := notifService.Send(ctx, e.ID, a.ID, notification.ChannelSMS, &notification.Message{
						To:   *a.Phone,
						Body: smsBody,
						Lang: e.Language,
					}); err != nil {
						logger.Error().Err(err).Str("attendee_id", a.ID).Msg("cancel notify: failed to send SMS")
					} else {
						sent++
					}
				}
			}

			logger.Info().
				Str("event_id", e.ID).
				Int("sent", sent).
				Msg("cancel notify: cancellation notifications dispatched")
		})
	}

	sched := scheduler.New(logger)
	reminderJob := scheduler.NewReminderJob(reminderStore, db, notifService, cfg.BaseURL, logger)
	reminderJob.SetTemplateResolver(messageTemplateService.Resolve)
	cleanupJob := scheduler.NewCleanupJob(db, logger)

	// Wire retention warning notifications into the cleanup job.
	if notifRegistry.Has(notification.ChannelEmail) {
		cleanupJob.SetRetentionNotify(func(ctx context.Context, organizerEmail, eventTitle string, expiresAt time.Time) {
			lang := "en"
			if organizer, err := authStore.FindOrganizerByEmail(ctx, organizerEmail); err == nil && organizer != nil {
				lang = organizer.Language
			}
			expiresStr := templates.FormatEventDate(expiresAt, lang)
			dashboardURL := cfg.BaseURL + "/events"

			subject, htmlBody, plainBody, err := templates.RenderRetentionWarning(lang, eventTitle, expiresStr, dashboardURL)
			if err != nil {
				logger.Error().Err(err).Str("event_title", eventTitle).Msg("retention warning: failed to render template")
				return
			}

			provider, provErr := notifRegistry.Get(notification.ChannelEmail)
			if provErr != nil {
				logger.Error().Err(provErr).Msg("retention warning: no email provider")
				return
			}

			if _, sendErr := provider.Send(ctx, &notification.Message{
				To:      organizerEmail,
				Subject: subject + " — " + eventTitle,
				Body:    htmlBody,
				Plain:   plainBody,
				Lang:    lang,
			}); sendErr != nil {
				logger.Error().Err(sendErr).Str("email", organizerEmail).Msg("retention warning: failed to send email")
			}
		})
	}

	// Clean up uploaded files when events are deleted by retention policy.
	cleanupJob.SetOnDeleteEvent(func(eventID string) {
		entries, err := os.ReadDir(uploadsDir)
		if err != nil {
			return
		}
		prefix := eventID + "_"
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
				_ = os.Remove(filepath.Join(uploadsDir, entry.Name()))
				logger.Debug().Str("file", entry.Name()).Msg("cleaned up uploaded file for deleted event")
			}
		}
	})

	sched.Register(reminderJob)
	sched.Register(cleanupJob)

	// Copy invite card when a new series occurrence is generated.
	seriesService.SetOnCreateOccurrence(func(ctx context.Context, seriesID, occurrenceID string) {
		events, err := eventStore.FindBySeriesID(ctx, seriesID)
		if err != nil || len(events) == 0 {
			return
		}
		for _, e := range events {
			if e.ID == occurrenceID {
				continue
			}
			card, err := inviteService.GetByEventID(ctx, e.ID)
			if err != nil || card == nil {
				continue
			}
			_, err = inviteService.Save(ctx, occurrenceID, invite.SaveInviteRequest{
				TemplateID:     card.TemplateID,
				Heading:        card.Heading,
				Body:           card.Body,
				Footer:         card.Footer,
				PrimaryColor:   card.PrimaryColor,
				SecondaryColor: card.SecondaryColor,
				Font:           card.Font,
				CustomData:     card.CustomData,
			})
			if err != nil {
				logger.Error().Err(err).
					Str("series_id", seriesID).
					Str("occurrence_id", occurrenceID).
					Msg("failed to copy invite card for series occurrence")
			}
			break
		}
	})

	// Register series generator background job.
	seriesJob := scheduler.NewSeriesGeneratorJob(seriesService, logger)
	sched.Register(seriesJob)

	// Wire up admin stats layer.
	statsStore := stats.NewStore(db)
	statsService := stats.NewService(statsStore, logger)
	adminMiddleware := auth.RequireAdmin()
	statsHandler := stats.NewHandler(statsService, authMiddleware, adminMiddleware, logger)

	// Wire up instance setup/config layer. DB-backed non-secret overrides
	// (instance name, default timezone, signups, support email) are overlaid
	// on top of the env-derived config at startup.
	instanceConfigStore := instanceconfig.NewStore(db)
	instanceConfigService := instanceconfig.NewService(instanceConfigStore)
	if overrides, err := instanceConfigStore.GetAll(context.Background()); err == nil {
		cfg.ApplyInstanceOverrides(overrides)
	}
	instanceConfigHandler := instanceconfig.NewHandler(instanceConfigService, authMiddleware, adminMiddleware, logger)

	s := &Server{
		cfg:                   cfg,
		db:                    db,
		logger:                logger,
		authHandler:           authHandler,
		eventHandler:          eventHandler,
		seriesHandler:         seriesHandler,
		rsvpHandler:           rsvpHandler,
		inviteHandler:         inviteHandler,
		messageHandler:        messageHandler,
		messageTemplateHandler: messageTemplateHandler,
		questionHandler:       questionHandler,
		feedbackHandler:       feedbackHandler,
		reminderHandler:       reminderHandler,
		commentHandler:        commentHandler,
		webhookHandler:        webhookHandler,
		notifHandler:          notifHandler,
		notifService:          notifService,
		statsHandler:          statsHandler,
		suppressionHandler:    suppressionHandler,
		instanceConfigHandler: instanceConfigHandler,
		scheduler:             sched,
		securityMw:            secMw,
		uploadsDir:            uploadsDir,
		smsEnabled:            smsEnabled,
	}

	router := s.routes()

	s.http = &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return s
}

// Start begins listening and blocks until the provided context is cancelled.
// It performs a graceful shutdown when the context is done.
func (s *Server) Start(ctx context.Context) error {
	// Start background scheduler.
	s.scheduler.Start(ctx)

	errCh := make(chan error, 1)

	go func() {
		s.logger.Info().Str("addr", s.http.Addr).Msg("server listening")
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.logger.Info().Msg("shutting down server")
	}

	// Stop scheduler first.
	s.scheduler.Stop()

	// Stop rate limiter cleanup goroutines.
	s.securityMw.AuthRateLimiter.Stop()
	s.securityMw.RSVPRateLimiter.Stop()
	s.securityMw.GeneralRateLimiter.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	s.logger.Info().Msg("server stopped gracefully")
	return nil
}

package messagetemplate

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/yannkr/openrsvp/internal/notification/templates"
)

// Field length limits.
const (
	maxSubjectLen = 200
	maxBodyLen    = 5000
)

// Service implements the message template business logic.
type Service struct {
	store  *Store
	logger zerolog.Logger
}

// NewService creates a new message template Service.
func NewService(store *Store, logger zerolog.Logger) *Service {
	return &Service{store: store, logger: logger}
}

// Resolve returns the effective subject/body for a message type on an event:
// the organizer's override if one is set, otherwise the language default.
func (s *Service) Resolve(ctx context.Context, eventID, messageType, lang string) (subject, body string, err error) {
	custom, err := s.store.FindByEventAndType(ctx, eventID, messageType)
	if err != nil {
		return "", "", fmt.Errorf("resolve message template: %w", err)
	}
	if custom != nil {
		return custom.Subject, custom.Body, nil
	}
	def := templates.DefaultFor(messageType, lang)
	return def.Subject, def.Body, nil
}

// ListEffective returns all message types for an event with their effective
// (override-or-default) content, for the organizer editor UI.
func (s *Service) ListEffective(ctx context.Context, eventID, lang string) ([]*EffectiveTemplate, error) {
	overrides, err := s.store.FindByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("list message templates: %w", err)
	}

	byType := make(map[string]*Template, len(overrides))
	for _, o := range overrides {
		byType[o.MessageType] = o
	}

	result := make([]*EffectiveTemplate, 0, len(AllTypes))
	for _, msgType := range AllTypes {
		if o, ok := byType[msgType]; ok {
			result = append(result, &EffectiveTemplate{
				MessageType:        msgType,
				Subject:            o.Subject,
				Body:               o.Body,
				IsCustomized:       true,
				AvailableVariables: AvailableVariablesFor(msgType),
			})
			continue
		}
		def := templates.DefaultFor(msgType, lang)
		result = append(result, &EffectiveTemplate{
			MessageType:        msgType,
			Subject:            def.Subject,
			Body:               def.Body,
			IsCustomized:       false,
			AvailableVariables: AvailableVariablesFor(msgType),
		})
	}

	return result, nil
}

// Upsert validates and stores an organizer's override for a message type.
func (s *Service) Upsert(ctx context.Context, eventID, messageType string, req *UpsertTemplateRequest) error {
	if !IsValidType(messageType) {
		return fmt.Errorf("invalid messageType: %s", messageType)
	}
	if req.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if len(req.Subject) > maxSubjectLen {
		return fmt.Errorf("subject must be %d characters or less", maxSubjectLen)
	}
	if req.Body == "" {
		return fmt.Errorf("body is required")
	}
	if len(req.Body) > maxBodyLen {
		return fmt.Errorf("body must be %d characters or less", maxBodyLen)
	}

	if err := s.store.Upsert(ctx, eventID, messageType, req.Subject, req.Body); err != nil {
		return fmt.Errorf("upsert message template: %w", err)
	}
	return nil
}

// ResetToDefault deletes the organizer's override, reverting to the language
// default for that message type.
func (s *Service) ResetToDefault(ctx context.Context, eventID, messageType string) error {
	if !IsValidType(messageType) {
		return fmt.Errorf("invalid messageType: %s", messageType)
	}
	if err := s.store.Delete(ctx, eventID, messageType); err != nil {
		return fmt.Errorf("reset message template: %w", err)
	}
	return nil
}

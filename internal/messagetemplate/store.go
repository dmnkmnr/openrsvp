package messagetemplate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yannkr/openrsvp/internal/database"
)

// Store handles database operations for per-event message template overrides.
type Store struct {
	db database.DB
}

// NewStore creates a new message template Store.
func NewStore(db database.DB) *Store {
	return &Store{db: db}
}

// FindByEventAndType retrieves the organizer's override for a message type on
// an event, or nil if none has been set (the caller should fall back to the
// language default).
func (s *Store) FindByEventAndType(ctx context.Context, eventID, messageType string) (*Template, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, event_id, message_type, subject, body, updated_at
		 FROM event_message_templates WHERE event_id = ? AND message_type = ?`,
		eventID, messageType,
	)
	return scanTemplate(row)
}

// FindByEvent retrieves all organizer overrides set for an event.
func (s *Store) FindByEvent(ctx context.Context, eventID string) ([]*Template, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, message_type, subject, body, updated_at
		 FROM event_message_templates WHERE event_id = ?`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("find message templates by event: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []*Template
	for rows.Next() {
		t, err := scanTemplateRow(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message templates: %w", err)
	}

	return templates, nil
}

// Upsert creates or replaces the organizer's override for a message type on
// an event.
func (s *Store) Upsert(ctx context.Context, eventID, messageType, subject, body string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.Must(uuid.NewV7()).String()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_message_templates (id, event_id, message_type, subject, body, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (event_id, message_type) DO UPDATE SET subject = excluded.subject, body = excluded.body, updated_at = excluded.updated_at`,
		id, eventID, messageType, subject, body, now,
	)
	if err != nil {
		return fmt.Errorf("upsert message template: %w", err)
	}
	return nil
}

// Delete removes the organizer's override for a message type on an event,
// resetting it back to the language default.
func (s *Store) Delete(ctx context.Context, eventID, messageType string) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM event_message_templates WHERE event_id = ? AND message_type = ?",
		eventID, messageType,
	)
	if err != nil {
		return fmt.Errorf("delete message template: %w", err)
	}
	return nil
}

// DeleteByEvent removes all overrides for an event. Not required by the FK's
// ON DELETE CASCADE when the event itself is deleted, but useful if an event
// is ever soft-reset independent of deletion.
func (s *Store) DeleteByEvent(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_message_templates WHERE event_id = ?", eventID)
	if err != nil {
		return fmt.Errorf("delete message templates by event: %w", err)
	}
	return nil
}

func scanTemplate(row *sql.Row) (*Template, error) {
	var t Template
	var updatedAt string

	err := row.Scan(&t.ID, &t.EventID, &t.MessageType, &t.Subject, &t.Body, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan message template: %w", err)
	}

	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &t, nil
}

func scanTemplateRow(rows *sql.Rows) (*Template, error) {
	var t Template
	var updatedAt string

	err := rows.Scan(&t.ID, &t.EventID, &t.MessageType, &t.Subject, &t.Body, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan message template row: %w", err)
	}

	t.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &t, nil
}

package messagetemplate

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/yannkr/openrsvp/internal/errcode"
)

// OrganizerFromCtx extracts the organizer ID from the request context.
type OrganizerFromCtx func(ctx context.Context) (id string, ok bool)

// EventOwnershipChecker verifies that the given organizer owns the event and
// returns the event's guest language. A non-nil error means the event was
// not found or is not owned by the organizer.
type EventOwnershipChecker func(ctx context.Context, eventID, organizerID string) (lang string, err error)

// Handler holds HTTP handlers for message template endpoints.
type Handler struct {
	service         *Service
	authMiddleware  func(http.Handler) http.Handler
	organizerFrom   OrganizerFromCtx
	checkEventOwner EventOwnershipChecker
	logger          zerolog.Logger
}

// NewHandler creates a new message template Handler.
func NewHandler(
	service *Service,
	authMiddleware func(http.Handler) http.Handler,
	organizerFrom OrganizerFromCtx,
	checkEventOwner EventOwnershipChecker,
	logger zerolog.Logger,
) *Handler {
	return &Handler{
		service:         service,
		authMiddleware:  authMiddleware,
		organizerFrom:   organizerFrom,
		checkEventOwner: checkEventOwner,
		logger:          logger,
	}
}

// Routes returns a chi.Router with all message template routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(auth chi.Router) {
		auth.Use(h.authMiddleware)
		auth.Get("/event/{eventId}", h.handleList)
		auth.Put("/event/{eventId}/{messageType}", h.handleUpsert)
		auth.Delete("/event/{eventId}/{messageType}", h.handleReset)
	})

	return r
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := h.organizerFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	eventID := chi.URLParam(r, "eventId")

	lang, err := h.checkEventOwner(r.Context(), eventID, organizerID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "event not found")
		return
	}

	list, err := h.service.ListEffective(r.Context(), eventID, lang)
	if err != nil {
		ref := errcode.Ref()
		h.logger.Error().Err(err).Str("error_ref", ref).Str("event_id", eventID).Msg("failed to list message templates")
		writeError(w, http.StatusInternalServerError, "internal_error", "an internal error occurred (ref: "+ref+")")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": list})
}

func (h *Handler) handleUpsert(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := h.organizerFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	eventID := chi.URLParam(r, "eventId")
	messageType := chi.URLParam(r, "messageType")

	if _, err := h.checkEventOwner(r.Context(), eventID, organizerID); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "event not found")
		return
	}

	var req UpsertTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	if err := h.service.Upsert(r.Context(), eventID, messageType, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
}

func (h *Handler) handleReset(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := h.organizerFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	eventID := chi.URLParam(r, "eventId")
	messageType := chi.URLParam(r, "messageType")

	if _, err := h.checkEventOwner(r.Context(), eventID, organizerID); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "event not found")
		return
	}

	if err := h.service.ResetToDefault(r.Context(), eventID, messageType); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "ok"}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   errCode,
		"message": message,
	})
}

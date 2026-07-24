package messagetemplate

import "time"

// Message type constants identify which guest notification a template
// override applies to. These string values are also used as the [messageType]
// key into internal/notification/templates' embedded language defaults, so
// they must stay in sync with the filenames under
// internal/notification/templates/defaults/.
const (
	TypeRSVPConfirmation  = "rsvp_confirmation"
	TypeCancellation      = "cancellation"
	TypeReminder          = "reminder"
	TypeWaitlistPromotion = "waitlist_promotion"
	TypeImportInvite      = "import_invite"
)

// AllTypes lists every customizable guest notification message type, in the
// order they should be presented to the organizer.
var AllTypes = []string{
	TypeRSVPConfirmation,
	TypeCancellation,
	TypeReminder,
	TypeWaitlistPromotion,
	TypeImportInvite,
}

// availableVariables lists the interpolation placeholders available for each
// message type, returned to the frontend as editing hints.
var availableVariables = map[string][]string{
	TypeRSVPConfirmation:  {"guestName", "eventTitle", "eventDate", "location", "rsvpStatus", "rsvpLink"},
	TypeCancellation:      {"guestName", "eventTitle", "eventDate", "location"},
	TypeReminder:          {"guestName", "eventTitle", "eventDate", "location", "rsvpLink"},
	TypeWaitlistPromotion: {"guestName", "eventTitle", "eventDate", "location", "rsvpLink"},
	TypeImportInvite:      {"guestName", "eventTitle", "eventDate", "location", "rsvpLink"},
}

// AvailableVariablesFor returns the interpolation placeholders available for
// a message type, or nil if the type is unknown.
func AvailableVariablesFor(messageType string) []string {
	return availableVariables[messageType]
}

// IsValidType reports whether the given string is one of the known message
// types.
func IsValidType(messageType string) bool {
	for _, t := range AllTypes {
		if t == messageType {
			return true
		}
	}
	return false
}

// Template is an organizer's customization of a guest notification's
// subject/body for one event. A row only exists when the organizer has
// customized that message type; its absence means "use the language
// default".
type Template struct {
	ID          string    `json:"id"`
	EventID     string    `json:"eventId"`
	MessageType string    `json:"messageType"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// UpsertTemplateRequest is the request body for setting a template override.
type UpsertTemplateRequest struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EffectiveTemplate is the resolved subject/body for a message type,
// combined with metadata about whether it's an organizer override or a
// language default. This is what the editor UI lists and prefills from.
type EffectiveTemplate struct {
	MessageType        string   `json:"messageType"`
	Subject            string   `json:"subject"`
	Body               string   `json:"body"`
	IsCustomized       bool     `json:"isCustomized"`
	AvailableVariables []string `json:"availableVariables"`
}

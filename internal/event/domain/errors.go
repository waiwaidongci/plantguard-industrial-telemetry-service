package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrMessageRequired     = shareddomain.New("event_message_required", "event message is required", 400)
	ErrAlreadyAcknowledged = shareddomain.New("event_already_acknowledged", "event has already been acknowledged", 409)
)

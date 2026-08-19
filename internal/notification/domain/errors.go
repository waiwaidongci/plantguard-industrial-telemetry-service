package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrUnsupportedChannel = shareddomain.New("unsupported_notification_channel", "notification channel is not supported", 400)
	ErrSubjectRequired    = shareddomain.New("notification_subject_required", "notification subject is required", 400)
)

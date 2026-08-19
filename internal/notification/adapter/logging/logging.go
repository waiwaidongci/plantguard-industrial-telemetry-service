package logging

import (
	"context"
	"log/slog"

	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
)

type LoggingSender struct {
	logger *slog.Logger
}

func NewLoggingSender(logger *slog.Logger) *LoggingSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingSender{logger: logger}
}

func (s *LoggingSender) Name() string {
	return "log"
}

func (s *LoggingSender) Send(_ context.Context, notification notificationdomain.Notification) error {
	s.logger.Info("notification_sent",
		"notification_id", notification.ID,
		"tenant_id", notification.TenantID,
		"channel", notification.Channel,
		"subject", notification.Subject,
	)
	return nil
}

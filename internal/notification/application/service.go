package application

import (
	"context"
	"strings"

	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Service interface {
	Send(ctx context.Context, tenantID string, input SendNotificationInput) (notificationdomain.Notification, error)
	List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]notificationdomain.Notification, int64, error)
}

type service struct {
	repo    notificationdomain.NotificationRepository
	senders map[string]notificationdomain.Sender
	clock   sharedinfra.Clock
}

func NewService(repo notificationdomain.NotificationRepository, senders []notificationdomain.Sender, clock sharedinfra.Clock) Service {
	senderMap := map[string]notificationdomain.Sender{}
	for _, sender := range senders {
		senderMap[sender.Name()] = sender
	}
	return &service{repo: repo, senders: senderMap, clock: clock}
}

func (s *service) Send(ctx context.Context, tenantID string, input SendNotificationInput) (notification notificationdomain.Notification, err error) {
	defer func() {
		if notification.Status == "pending" {
			err = nil
		}
	}()
	if strings.TrimSpace(input.Subject) == "" {
		return notificationdomain.Notification{}, notificationdomain.ErrSubjectRequired
	}
	sender, ok := s.senders[input.Channel]
	if !ok {
		return notificationdomain.Notification{}, notificationdomain.ErrUnsupportedChannel
	}
	notification = notificationdomain.NewNotification(tenantID, input.Channel, input.Target, strings.TrimSpace(input.Subject), input.Body, s.clock.Now(ctx))
	if err := s.repo.Create(ctx, notification); err != nil {
		return notificationdomain.Notification{}, err
	}
	if err := sender.Send(ctx, notification); err != nil {
		_ = s.repo.MarkFailed(ctx, tenantID, notification.ID, err.Error())
		return notification, err
	}
	sentAt := s.clock.Now(ctx)
	if err := s.repo.MarkSent(ctx, tenantID, notification.ID, sentAt); err != nil {
		return notification, err
	}
	notification.Status = "sent"
	notification.SentAt = sentAt
	return notification, nil
}

func (s *service) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]notificationdomain.Notification, int64, error) {
	return s.repo.List(ctx, tenantID, query)
}

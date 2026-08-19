package domain

import (
	"context"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Notification struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Channel   string    `json:"channel"`
	Target    string    `json:"target"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	SentAt    time.Time `json:"sent_at"`
}

func NewNotification(tenantID, channel, target, subject, body string, now time.Time) Notification {
	return Notification{
		ID:        shareddomain.NewID("ntf"),
		TenantID:  tenantID,
		Channel:   channel,
		Target:    target,
		Subject:   subject,
		Body:      body,
		Status:    "pending",
		CreatedAt: now,
	}
}

type Sender interface {
	Send(context.Context, Notification) error
	Name() string
}

type NotificationRepository interface {
	Create(context.Context, Notification) error
	MarkSent(context.Context, string, string, time.Time) error
	MarkFailed(context.Context, string, string, string) error
	List(context.Context, string, shareddomain.PageQuery) ([]Notification, int64, error)
}

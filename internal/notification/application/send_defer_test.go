package application

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

var errMarkSent = errors.New("mark sent failed")

type failingMarkSentRepo struct{}

func (failingMarkSentRepo) Create(context.Context, notificationdomain.Notification) error { return nil }
func (failingMarkSentRepo) MarkSent(context.Context, string, string, time.Time) error {
	return errMarkSent
}
func (failingMarkSentRepo) MarkFailed(context.Context, string, string, string) error { return nil }
func (failingMarkSentRepo) List(context.Context, string, shareddomain.PageQuery) ([]notificationdomain.Notification, int64, error) {
	return nil, 0, nil
}

type okSender struct{}

func (okSender) Name() string { return "log" }
func (okSender) Send(context.Context, notificationdomain.Notification) error {
	return nil
}

type failingSender struct{}

func (failingSender) Name() string { return "log" }
func (failingSender) Send(context.Context, notificationdomain.Notification) error {
	return errMarkSent
}

type sendClock struct{}

func (sendClock) Now(context.Context) time.Time { return time.Unix(1700000000, 0).UTC() }

func TestSendPropagatesMarkSentError(t *testing.T) {
	svc := NewService(failingMarkSentRepo{}, []notificationdomain.Sender{okSender{}}, sendClock{})
	_, err := svc.Send(context.Background(), "tenant-1", SendNotificationInput{
		Channel: "log",
		Subject: "subject",
		Body:    "body",
	})
	if !errors.Is(err, errMarkSent) {
		t.Fatalf("expected MarkSent error to be preserved, got %v", err)
	}
}

func TestSendPropagatesSenderError(t *testing.T) {
	svc := NewService(failingMarkSentRepo{}, []notificationdomain.Sender{failingSender{}}, sendClock{})
	_, err := svc.Send(context.Background(), "tenant-1", SendNotificationInput{
		Channel: "log",
		Subject: "subject",
		Body:    "body",
	})
	if !errors.Is(err, errMarkSent) {
		t.Fatalf("expected sender error to be preserved, got %v", err)
	}
}

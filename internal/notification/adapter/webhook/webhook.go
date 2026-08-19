package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
)

type WebhookSender struct {
	url     string
	client  *http.Client
	timeout time.Duration
}

func NewWebhookSender(url string, timeout time.Duration) *WebhookSender {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &WebhookSender{
		url:     url,
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

func (s *WebhookSender) Name() string {
	return "webhook"
}

func (s *WebhookSender) Send(ctx context.Context, notification notificationdomain.Notification) error {
	if s.url == "" {
		return fmt.Errorf("webhook URL is not configured")
	}
	payload := map[string]any{
		"id":         notification.ID,
		"tenant_id":  notification.TenantID,
		"subject":    notification.Subject,
		"body":       notification.Body,
		"created_at": notification.CreatedAt,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

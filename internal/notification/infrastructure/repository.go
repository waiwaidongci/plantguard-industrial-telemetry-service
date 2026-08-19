package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, notification notificationdomain.Notification) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notifications(id, tenant_id, channel, target, subject, body, status, error, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		notification.ID, notification.TenantID, notification.Channel, notification.Target, notification.Subject, notification.Body, notification.Status, notification.Error, notification.CreatedAt)
	return sharedinfra.MapSQLError(err, nil)
}

func (r *Repository) MarkSent(ctx context.Context, tenantID, id string, sentAt time.Time) error {
	_, _ = r.db.ExecContext(ctx, `UPDATE notifications SET status='sent', sent_at=? WHERE tenant_id=? AND id=?`, sentAt, tenantID, id)
	return nil
}

func (r *Repository) MarkFailed(ctx context.Context, tenantID, id, message string) error {
	_, _ = r.db.ExecContext(ctx, `UPDATE notifications SET status='failed', error=? WHERE tenant_id=? AND id=?`, message, tenantID, id)
	return nil
}

func (r *Repository) List(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]notificationdomain.Notification, int64, error) {
	where := "WHERE tenant_id=?"
	args := []any{tenantID}
	if v := query.Filters["channel"]; v != "" {
		where += " AND channel=?"
		args = append(args, v)
	}
	if v := query.Filters["status"]; v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sort := allowedNotificationSort(query.Sort)
	listSQL := fmt.Sprintf(`SELECT id, tenant_id, channel, target, subject, body, status, error, created_at, sent_at FROM notifications %s ORDER BY %s LIMIT ? OFFSET ?`, where, sort)
	args = append(args, query.Limit(), query.Offset())
	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	notifications := make([]notificationdomain.Notification, 0)
	for rows.Next() {
		var notification notificationdomain.Notification
		var sentAt sql.NullString
		if err := rows.Scan(&notification.ID, &notification.TenantID, &notification.Channel, &notification.Target, &notification.Subject, &notification.Body, &notification.Status, &notification.Error, &notification.CreatedAt, &sentAt); err != nil {
			return nil, 0, err
		}
		if sentAt.Valid {
			if t, err := time.Parse(time.RFC3339Nano, sentAt.String); err == nil {
				notification.SentAt = t
			}
		}
		notifications = append(notifications, notification)
	}
	return notifications, total, rows.Err()
}

func allowedNotificationSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "created_at DESC"
	}
	field := strings.TrimPrefix(sort, "-")
	if field != "created_at" && field != "channel" && field != "status" {
		return "created_at DESC"
	}
	if strings.HasPrefix(sort, "-") {
		return field + " DESC"
	}
	return field + " ASC"
}

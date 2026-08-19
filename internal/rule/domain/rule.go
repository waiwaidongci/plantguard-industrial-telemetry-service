package domain

import (
	"context"
	"encoding/json"
	"time"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type Rule struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Metric          string            `json:"metric"`
	Condition       string            `json:"condition"`
	Threshold       float64           `json:"threshold"`
	DurationSeconds int               `json:"duration_seconds"`
	WindowSeconds   int               `json:"window_seconds"`
	Aggregation     string            `json:"aggregation"`
	TagFilters      map[string]string `json:"tag_filters"`
	Severity        string            `json:"severity"`
	Enabled         bool              `json:"enabled"`
	Actions         []string          `json:"actions"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Version         int64             `json:"version"`
}

func NewRule(tenantID, name, metric, condition string, threshold float64, severity string, now time.Time) Rule {
	return Rule{
		ID:            shareddomain.NewID("rul"),
		TenantID:      tenantID,
		Name:          name,
		Metric:        metric,
		Condition:     condition,
		Threshold:     threshold,
		Aggregation:   "avg",
		WindowSeconds: 60,
		Severity:      severity,
		Enabled:       true,
		TagFilters:    map[string]string{},
		Actions:       []string{"log"},
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

func (r Rule) MarshalTagFilters() (string, error) {
	raw, err := json.Marshal(r.TagFilters)
	return string(raw), err
}

func (r Rule) MarshalActions() (string, error) {
	raw, err := json.Marshal(r.Actions)
	return string(raw), err
}

func UnmarshalTagFilters(raw string) map[string]string {
	filters := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &filters)
	return filters
}

func UnmarshalActions(raw string) []string {
	actions := []string{}
	_ = json.Unmarshal([]byte(raw), &actions)
	return actions
}

type RuleRepository interface {
	Create(context.Context, Rule) error
	GetByID(context.Context, string, string) (Rule, error)
	Update(context.Context, Rule) error
	List(context.Context, string, shareddomain.PageQuery) ([]Rule, int64, error)
	ListEnabledByTenant(context.Context, string) ([]Rule, error)
}

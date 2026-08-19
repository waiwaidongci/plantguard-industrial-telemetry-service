package application

import (
	"context"
	"strings"
	"time"

	ruledomain "github.com/acme/plantguard/internal/rule/domain"
	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Service interface {
	CreateRule(ctx context.Context, tenantID, idempotencyKey string, input CreateRuleInput) (ruledomain.Rule, bool, error)
	GetRule(ctx context.Context, tenantID, id string) (ruledomain.Rule, error)
	UpdateRule(ctx context.Context, tenantID, id string, input UpdateRuleInput) (ruledomain.Rule, error)
	ListRules(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]ruledomain.Rule, int64, error)
	ListEnabled(ctx context.Context, tenantID string) ([]ruledomain.Rule, error)
	Evaluate(ctx context.Context, rule ruledomain.Rule, facts ruledomain.DeviceFacts, value float64, occurredAt time.Time) (ruledomain.Violation, bool)
}

type service struct {
	repo        ruledomain.RuleRepository
	idempotency sharedapplication.IdempotencyStore
	clock       sharedinfra.Clock
	evaluator   ruledomain.Evaluator
}

func NewService(repo ruledomain.RuleRepository, idempotency sharedapplication.IdempotencyStore, clock sharedinfra.Clock) Service {
	return &service{repo: repo, idempotency: idempotency, clock: clock, evaluator: ruledomain.NewEvaluator()}
}

func (s *service) CreateRule(ctx context.Context, tenantID, idempotencyKey string, input CreateRuleInput) (ruledomain.Rule, bool, error) {
	if strings.TrimSpace(input.Name) == "" {
		return ruledomain.Rule{}, false, ruledomain.ErrRuleNameRequired
	}
	if strings.TrimSpace(input.Metric) == "" {
		return ruledomain.Rule{}, false, ruledomain.ErrMetricRequired
	}
	if !validCondition(input.Condition) {
		return ruledomain.Rule{}, false, ruledomain.ErrConditionInvalid
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return ruledomain.Rule{}, false, err
		}
		if found {
			rule, err := s.repo.GetByID(ctx, tenantID, record.ResourceID)
			return rule, false, err
		}
	}
	now := s.clock.Now(ctx)
	rule := ruledomain.NewRule(tenantID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Metric), input.Condition, input.Threshold, normalizeSeverity(input.Severity), now)
	rule.Description = input.Description
	rule.DurationSeconds = input.DurationSeconds
	rule.WindowSeconds = input.WindowSeconds
	rule.Aggregation = normalizeAggregation(input.Aggregation)
	rule.TagFilters = input.TagFilters
	rule.Actions = normalizeActions(input.Actions)
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return ruledomain.Rule{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{TenantID: tenantID, Key: idempotencyKey, ResourceKind: "rule", ResourceID: rule.ID}); err != nil {
			return ruledomain.Rule{}, true, err
		}
	}
	return rule, true, nil
}

func (s *service) GetRule(ctx context.Context, tenantID, id string) (ruledomain.Rule, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *service) UpdateRule(ctx context.Context, tenantID, id string, input UpdateRuleInput) (ruledomain.Rule, error) {
	rule, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return ruledomain.Rule{}, err
	}
	if input.Version != 0 && input.Version != rule.Version {
		return ruledomain.Rule{}, shareddomain.ErrPrecondition
	}
	if strings.TrimSpace(input.Name) != "" {
		rule.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Description) != "" {
		rule.Description = input.Description
	}
	if strings.TrimSpace(input.Metric) != "" {
		rule.Metric = strings.TrimSpace(input.Metric)
	}
	if input.Condition != "" {
		if !validCondition(input.Condition) {
			return ruledomain.Rule{}, ruledomain.ErrConditionInvalid
		}
		rule.Condition = input.Condition
	}
	if input.Threshold != nil {
		rule.Threshold = *input.Threshold
	}
	if input.DurationSeconds != nil {
		rule.DurationSeconds = *input.DurationSeconds
	}
	if input.WindowSeconds != nil {
		rule.WindowSeconds = *input.WindowSeconds
	}
	if input.Aggregation != "" {
		rule.Aggregation = normalizeAggregation(input.Aggregation)
	}
	if input.TagFilters != nil {
		rule.TagFilters = input.TagFilters
	}
	if input.Severity != "" {
		rule.Severity = normalizeSeverity(input.Severity)
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.Actions != nil {
		rule.Actions = normalizeActions(input.Actions)
	}
	rule.UpdatedAt = s.clock.Now(ctx)
	rule.Version++
	if err := s.repo.Update(ctx, rule); err != nil {
		return ruledomain.Rule{}, err
	}
	return rule, nil
}

func (s *service) ListRules(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]ruledomain.Rule, int64, error) {
	return s.repo.List(ctx, tenantID, query)
}

func (s *service) ListEnabled(ctx context.Context, tenantID string) ([]ruledomain.Rule, error) {
	return s.repo.ListEnabledByTenant(ctx, tenantID)
}

func (s *service) Evaluate(_ context.Context, rule ruledomain.Rule, facts ruledomain.DeviceFacts, value float64, occurredAt time.Time) (ruledomain.Violation, bool) {
	return s.evaluator.Evaluate(rule, facts, value, occurredAt.UTC().Format(time.RFC3339Nano))
}

func validCondition(condition string) bool {
	switch condition {
	case "gt", "gte", "lt", "lte", "eq":
		return true
	default:
		return false
	}
}

func normalizeSeverity(severity string) string {
	switch severity {
	case "critical", "warning", "info":
		return severity
	default:
		return "warning"
	}
}

func normalizeAggregation(aggregation string) string {
	switch aggregation {
	case "avg", "min", "max", "last", "count":
		return aggregation
	default:
		return "avg"
	}
}

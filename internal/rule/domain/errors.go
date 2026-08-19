package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrRuleNameRequired = shareddomain.New("rule_name_required", "rule name is required", 400)
	ErrMetricRequired   = shareddomain.New("rule_metric_required", "rule metric is required", 400)
	ErrConditionInvalid = shareddomain.New("rule_condition_invalid", "condition must be one of gt, gte, lt, lte, eq", 400)
)

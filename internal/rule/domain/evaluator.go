package domain

import (
	"fmt"
	"math"
	"strings"
)

type DeviceFacts struct {
	DeviceID string
	Tags     []string
}

type Violation struct {
	RuleID     string
	RuleName   string
	DeviceID   string
	Metric     string
	Value      float64
	Threshold  float64
	Condition  string
	Severity   string
	Message    string
	OccurredAt string
}

type Evaluator struct{}

func NewEvaluator() Evaluator {
	return Evaluator{}
}

func (Evaluator) Evaluate(rule Rule, facts DeviceFacts, value float64, occurredAt string) (Violation, bool) {
	if !tagsMatch(rule.TagFilters, facts.Tags) {
		return Violation{}, false
	}
	if !compare(rule.Condition, value, rule.Threshold) {
		return Violation{}, false
	}
	return Violation{
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		DeviceID:   facts.DeviceID,
		Metric:     rule.Metric,
		Value:      value,
		Threshold:  rule.Threshold,
		Condition:  rule.Condition,
		Severity:   rule.Severity,
		Message:    rule.MessageFor(value),
		OccurredAt: occurredAt,
	}, true
}

func compare(condition string, value, threshold float64) bool {
	switch strings.ToLower(condition) {
	case "gt", ">":
		return value > threshold
	case "gte", ">=":
		return value >= threshold
	case "lt", "<":
		return value < threshold
	case "lte", "<=":
		return value <= threshold
	case "eq", "=":
		return math.Abs(value-threshold) < 1e-9
	default:
		return value > threshold
	}
}

func tagsMatch(filters map[string]string, tags []string) bool {
	if len(filters) == 0 {
		return true
	}
	tagSet := map[string]string{}
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			tagSet[parts[0]] = parts[1]
		}
	}
	for key, value := range filters {
		if tagSet[key] != value {
			return false
		}
	}
	return true
}

func (r Rule) MessageFor(value float64) string {
	if r.Description != "" {
		return r.Description
	}
	return r.Metric + " " + r.Condition + " threshold " + formatFloat(r.Threshold) + ", observed " + formatFloat(value)
}

func formatFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

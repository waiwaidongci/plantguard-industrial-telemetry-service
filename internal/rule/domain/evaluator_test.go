package domain

import (
	"testing"
	"time"
)

func nowForTest() time.Time {
	return time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
}

func TestEvaluatorMatchesThresholdAndTags(t *testing.T) {
	rule := NewRule("tnt_1", "High Temp", "temperature", "gt", 80, "warning", nowForTest())
	rule.TagFilters = map[string]string{"zone": "a"}
	evaluator := NewEvaluator()

	violation, matched := evaluator.Evaluate(rule, DeviceFacts{
		DeviceID: "dev_1",
		Tags:     []string{"zone=a", "criticality=high"},
	}, 86.5, "2026-08-20T00:00:00Z")

	if !matched {
		t.Fatal("expected rule to match")
	}
	if violation.DeviceID != "dev_1" || violation.Value != 86.5 {
		t.Fatalf("unexpected violation: %#v", violation)
	}
}

func TestEvaluatorRejectsNonMatchingTag(t *testing.T) {
	rule := NewRule("tnt_1", "High Temp", "temperature", "gt", 80, "warning", nowForTest())
	rule.TagFilters = map[string]string{"zone": "b"}
	evaluator := NewEvaluator()

	_, matched := evaluator.Evaluate(rule, DeviceFacts{
		DeviceID: "dev_1",
		Tags:     []string{"zone=a"},
	}, 86.5, "2026-08-20T00:00:00Z")

	if matched {
		t.Fatal("expected tag filter to reject the rule")
	}
}

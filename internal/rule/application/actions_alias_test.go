package application

import (
	"context"
	"reflect"
	"testing"
	"time"

	ruledomain "github.com/acme/plantguard/internal/rule/domain"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

type captureRuleRepo struct {
	created ruledomain.Rule
}

func (r *captureRuleRepo) Create(_ context.Context, rule ruledomain.Rule) error { r.created = rule; return nil }
func (captureRuleRepo) GetByID(context.Context, string, string) (ruledomain.Rule, error) {
	return ruledomain.Rule{}, shareddomain.ErrNotFound
}
func (captureRuleRepo) Update(context.Context, ruledomain.Rule) error { return nil }
func (captureRuleRepo) List(context.Context, string, shareddomain.PageQuery) ([]ruledomain.Rule, int64, error) {
	return nil, 0, nil
}
func (captureRuleRepo) ListEnabledByTenant(context.Context, string) ([]ruledomain.Rule, error) {
	return nil, nil
}

type actionClock struct{}

func (actionClock) Now(context.Context) time.Time { return time.Unix(1700000000, 0).UTC() }

func TestCreateRuleDoesNotMutateActionInput(t *testing.T) {
	repo := &captureRuleRepo{}
	svc := NewService(repo, nil, actionClock{})
	original := []string{"log", "webhook", "invalid", "maintenance"}
	want := append([]string(nil), original...)
	input := CreateRuleInput{
		Name:      "pressure high",
		Metric:    "pressure",
		Condition: "gt",
		Threshold: 10,
		Actions:   original,
	}
	_, _, err := svc.CreateRule(context.Background(), "tenant-1", "", input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input.Actions, want) {
		t.Fatalf("input actions were mutated: %v -> %v", want, input.Actions)
	}
}

func TestRuleNormalizationFunctions(t *testing.T) {
	t.Run("severity", func(t *testing.T) {
		if got := normalizeSeverity("critical"); got != "critical" {
			t.Fatalf("expected critical, got %q", got)
		}
	})
	t.Run("aggregation", func(t *testing.T) {
		if got := normalizeAggregation("max"); got != "max" {
			t.Fatalf("expected max, got %q", got)
		}
	})
	t.Run("actions", func(t *testing.T) {
		actions := []string{"log", "webhook", "invalid", "maintenance"}
		got := normalizeActions(actions)
		if len(got) != 3 || got[0] != "log" || got[1] != "webhook" || got[2] != "maintenance" {
			t.Fatalf("unexpected normalized actions: %v", got)
		}
	})
}

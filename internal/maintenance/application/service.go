package application

import (
	"context"
	"strings"
	"time"

	maintenancedomain "github.com/acme/plantguard/internal/maintenance/domain"
	sharedapplication "github.com/acme/plantguard/internal/shared/application"
	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
)

type Service interface {
	CreatePlan(ctx context.Context, tenantID, idempotencyKey string, input CreatePlanInput) (maintenancedomain.Plan, bool, error)
	GetPlan(ctx context.Context, tenantID, id string) (maintenancedomain.Plan, error)
	UpdatePlan(ctx context.Context, tenantID, id string, input UpdatePlanInput) (maintenancedomain.Plan, error)
	ListPlans(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Plan, int64, error)
	GetTask(ctx context.Context, tenantID, id string) (maintenancedomain.Task, error)
	TransitionTask(ctx context.Context, tenantID, id string, input TransitionTaskInput) (maintenancedomain.Task, error)
	ListTasks(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Task, int64, error)
	CreateDueTasks(ctx context.Context, now time.Time) ([]maintenancedomain.Task, error)
}

type service struct {
	plans       maintenancedomain.PlanRepository
	tasks       maintenancedomain.TaskRepository
	idempotency sharedapplication.IdempotencyStore
	clock       sharedinfra.Clock
}

func NewService(plans maintenancedomain.PlanRepository, tasks maintenancedomain.TaskRepository, idempotency sharedapplication.IdempotencyStore, clock sharedinfra.Clock) Service {
	return &service{plans: plans, tasks: tasks, idempotency: idempotency, clock: clock}
}

func (s *service) CreatePlan(ctx context.Context, tenantID, idempotencyKey string, input CreatePlanInput) (maintenancedomain.Plan, bool, error) {
	if strings.TrimSpace(input.Name) == "" {
		return maintenancedomain.Plan{}, false, maintenancedomain.ErrPlanNameRequired
	}
	if !validTrigger(input.TriggerType) {
		return maintenancedomain.Plan{}, false, maintenancedomain.ErrInvalidTrigger
	}
	if idempotencyKey != "" {
		record, found, err := s.idempotency.Get(ctx, tenantID, idempotencyKey)
		if err != nil {
			return maintenancedomain.Plan{}, false, err
		}
		if found {
			plan, err := s.plans.GetByID(ctx, tenantID, record.ResourceID)
			return plan, false, err
		}
	}
	plan := maintenancedomain.NewPlan(tenantID, input.DeviceID, strings.TrimSpace(input.Name), input.TriggerType, input.IntervalHours, input.CalendarSpec, input.EventType, s.clock.Now(ctx))
	if err := s.plans.Create(ctx, plan); err != nil {
		return maintenancedomain.Plan{}, false, err
	}
	if idempotencyKey != "" {
		if err := s.idempotency.Put(ctx, sharedapplication.IdempotencyRecord{TenantID: tenantID, Key: idempotencyKey, ResourceKind: "maintenance_plan", ResourceID: plan.ID}); err != nil {
			return maintenancedomain.Plan{}, true, err
		}
	}
	return plan, true, nil
}

func (s *service) GetPlan(ctx context.Context, tenantID, id string) (maintenancedomain.Plan, error) {
	return s.plans.GetByID(ctx, tenantID, id)
}

func (s *service) UpdatePlan(ctx context.Context, tenantID, id string, input UpdatePlanInput) (maintenancedomain.Plan, error) {
	plan, err := s.plans.GetByID(ctx, tenantID, id)
	if err != nil {
		return maintenancedomain.Plan{}, err
	}
	if input.Version != 0 && input.Version != plan.Version {
		return maintenancedomain.Plan{}, shareddomain.ErrPrecondition
	}
	if strings.TrimSpace(input.Name) != "" {
		plan.Name = strings.TrimSpace(input.Name)
	}
	if input.DeviceID != "" {
		plan.DeviceID = input.DeviceID
	}
	if input.Status != "" {
		if input.Status != "active" && input.Status != "paused" {
			return maintenancedomain.Plan{}, shareddomain.New("invalid_plan_status", "plan status must be active or paused", 400)
		}
		plan.Status = input.Status
	}
	if input.IntervalHours != nil {
		plan.IntervalHours = *input.IntervalHours
		plan.NextDueAt = s.clock.Now(ctx).Add(time.Duration(*input.IntervalHours) * time.Hour)
	}
	if input.CalendarSpec != "" {
		plan.CalendarSpec = input.CalendarSpec
	}
	if input.EventType != "" {
		plan.EventType = input.EventType
	}
	plan.UpdatedAt = s.clock.Now(ctx)
	plan.Version++
	if err := s.plans.Update(ctx, plan); err != nil {
		return maintenancedomain.Plan{}, err
	}
	return plan, nil
}

func (s *service) ListPlans(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Plan, int64, error) {
	return s.plans.List(ctx, tenantID, query)
}

func (s *service) GetTask(ctx context.Context, tenantID, id string) (maintenancedomain.Task, error) {
	return s.tasks.GetByID(ctx, tenantID, id)
}

func (s *service) TransitionTask(ctx context.Context, tenantID, id string, input TransitionTaskInput) (maintenancedomain.Task, error) {
	task, err := s.tasks.GetByID(ctx, tenantID, id)
	if err != nil {
		return maintenancedomain.Task{}, err
	}
	if input.Version != 0 && input.Version != task.Version {
		return maintenancedomain.Task{}, shareddomain.ErrPrecondition
	}
	if !task.CanTransitionTo(input.Status) {
		return maintenancedomain.Task{}, maintenancedomain.ErrInvalidTransition
	}
	now := s.clock.Now(ctx)
	task.Status = input.Status
	task.Notes = input.Notes
	task.UpdatedAt = now
	if input.Status == "completed" {
		task.CompletedAt = now
	}
	task.Version++
	if err := s.tasks.Update(ctx, task); err != nil {
		return maintenancedomain.Task{}, err
	}
	return task, nil
}

func (s *service) ListTasks(ctx context.Context, tenantID string, query shareddomain.PageQuery) ([]maintenancedomain.Task, int64, error) {
	return s.tasks.List(ctx, tenantID, query)
}

func (s *service) CreateDueTasks(ctx context.Context, now time.Time) ([]maintenancedomain.Task, error) {
	plans, err := s.plans.ListActiveDue(ctx, now)
	if err != nil {
		return nil, err
	}
	created := make([]maintenancedomain.Task, 0)
	for _, plan := range plans {
		task := maintenancedomain.NewTask(plan.TenantID, plan.ID, plan.DeviceID, plan.NextDueAt, now)
		if err := s.tasks.Create(ctx, task); err != nil {
			return created, err
		}
		plan.NextDueAt = now.Add(time.Duration(plan.IntervalHours) * time.Hour)
		plan.UpdatedAt = now
		plan.Version++
		if err := s.plans.Update(ctx, plan); err != nil {
			return created, err
		}
		created = append(created, task)
	}
	return created, nil
}

func validTrigger(trigger string) bool {
	switch trigger {
	case "running_hours", "calendar", "event":
		return true
	default:
		return false
	}
}

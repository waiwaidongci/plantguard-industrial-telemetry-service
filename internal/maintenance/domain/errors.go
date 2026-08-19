package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrPlanNameRequired  = shareddomain.New("maintenance_plan_name_required", "maintenance plan name is required", 400)
	ErrInvalidTrigger    = shareddomain.New("invalid_maintenance_trigger", "trigger_type must be running_hours, calendar, or event", 400)
	ErrInvalidTransition = shareddomain.New("invalid_task_transition", "maintenance task cannot transition to the requested status", 409)
)

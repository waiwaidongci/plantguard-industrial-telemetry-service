package application

type CreatePlanInput struct {
	DeviceID      string `json:"device_id"`
	Name          string `json:"name"`
	TriggerType   string `json:"trigger_type"`
	IntervalHours int    `json:"interval_hours"`
	CalendarSpec  string `json:"calendar_spec"`
	EventType     string `json:"event_type"`
}

type UpdatePlanInput struct {
	DeviceID      string `json:"device_id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	IntervalHours *int   `json:"interval_hours,omitempty"`
	CalendarSpec  string `json:"calendar_spec"`
	EventType     string `json:"event_type"`
	Version       int64  `json:"version"`
}

type TransitionTaskInput struct {
	Status  string `json:"status"`
	Notes   string `json:"notes"`
	Version int64  `json:"version"`
}

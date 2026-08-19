package application

type CreateEventInput struct {
	DeviceID   string         `json:"device_id"`
	RuleID     string         `json:"rule_id"`
	Type       string         `json:"type"`
	Severity   string         `json:"severity"`
	Message    string         `json:"message"`
	Data       map[string]any `json:"data"`
	OccurredAt string         `json:"occurred_at"`
}

type AcknowledgeEventInput struct {
	AcknowledgedBy string `json:"acknowledged_by"`
	Version        int64  `json:"version"`
}

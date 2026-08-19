package application

type CreateRuleInput struct {
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
	Enabled         *bool             `json:"enabled,omitempty"`
	Actions         []string          `json:"actions"`
}

type UpdateRuleInput struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Metric          string            `json:"metric"`
	Condition       string            `json:"condition"`
	Threshold       *float64          `json:"threshold,omitempty"`
	DurationSeconds *int              `json:"duration_seconds,omitempty"`
	WindowSeconds   *int              `json:"window_seconds,omitempty"`
	Aggregation     string            `json:"aggregation"`
	TagFilters      map[string]string `json:"tag_filters"`
	Severity        string            `json:"severity"`
	Enabled         *bool             `json:"enabled,omitempty"`
	Actions         []string          `json:"actions"`
	Version         int64             `json:"version"`
}

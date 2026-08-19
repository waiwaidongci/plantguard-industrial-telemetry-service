package application

import (
	"time"

	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type IngestBatchInput struct {
	Source   string                    `json:"source"`
	Readings []telemetrydomain.Reading `json:"readings"`
}

type SummaryQuery struct {
	Since time.Time
	Until time.Time
}

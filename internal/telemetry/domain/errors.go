package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrBatchEmpty      = shareddomain.New("telemetry_batch_empty", "telemetry batch must contain at least one reading", 400)
	ErrDeviceDisabled  = shareddomain.New("device_disabled", "telemetry is not accepted for disabled devices", 409)
	ErrInvalidReading  = shareddomain.New("invalid_reading", "one or more telemetry readings are invalid", 400)
	ErrMissingReadings = shareddomain.New("missing_required_readings", "required metric readings are missing", 400)
)

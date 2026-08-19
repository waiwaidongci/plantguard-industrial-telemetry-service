package domain

import shareddomain "github.com/acme/plantguard/internal/shared/domain"

var (
	ErrDeviceNameRequired = shareddomain.New("device_name_required", "device name is required", 400)
	ErrModelNameRequired  = shareddomain.New("model_name_required", "model name is required", 400)
	ErrInvalidProtocol    = shareddomain.New("invalid_protocol", "protocol must be one of mqtt, http, modbus, opcua", 400)
)

package app

import (
	"database/sql"
	"log/slog"

	deviceapplication "github.com/acme/plantguard/internal/device/application"
	deviceinfra "github.com/acme/plantguard/internal/device/infrastructure"
	eventapplication "github.com/acme/plantguard/internal/event/application"
	eventinfra "github.com/acme/plantguard/internal/event/infrastructure"
	maintenanceapplication "github.com/acme/plantguard/internal/maintenance/application"
	maintenanceinfra "github.com/acme/plantguard/internal/maintenance/infrastructure"
	loggingadapter "github.com/acme/plantguard/internal/notification/adapter/logging"
	webhookadapter "github.com/acme/plantguard/internal/notification/adapter/webhook"
	notificationapplication "github.com/acme/plantguard/internal/notification/application"
	notificationdomain "github.com/acme/plantguard/internal/notification/domain"
	notificationinfra "github.com/acme/plantguard/internal/notification/infrastructure"
	ruleapplication "github.com/acme/plantguard/internal/rule/application"
	ruleinfra "github.com/acme/plantguard/internal/rule/infrastructure"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetryapplication "github.com/acme/plantguard/internal/telemetry/application"
	telemetryinfra "github.com/acme/plantguard/internal/telemetry/infrastructure"
	tenantapplication "github.com/acme/plantguard/internal/tenant/application"
	tenantinfra "github.com/acme/plantguard/internal/tenant/infrastructure"
)

type Repositories struct {
	Tenants       *tenantinfra.Repository
	Sites         *tenantinfra.SiteRepository
	Devices       *deviceinfra.DeviceRepository
	Models        *deviceinfra.ModelRepository
	Telemetry     *telemetryinfra.Repository
	Rules         *ruleinfra.Repository
	Events        *eventinfra.Repository
	Plans         *maintenanceinfra.PlanRepository
	Tasks         *maintenanceinfra.TaskRepository
	Notifications *notificationinfra.Repository
	Idempotency   *sharedinfra.IdempotencyStore
}

type Services struct {
	Tenant       tenantapplication.Service
	Device       deviceapplication.Service
	Telemetry    telemetryapplication.Service
	Event        eventapplication.Service
	Rule         ruleapplication.Service
	Maintenance  maintenanceapplication.Service
	Notification notificationapplication.Service
}

func NewRepositories(db *sql.DB) Repositories {
	return Repositories{
		Tenants:       tenantinfra.NewRepository(db),
		Sites:         tenantinfra.NewSiteRepository(db),
		Devices:       deviceinfra.NewDeviceRepository(db),
		Models:        deviceinfra.NewModelRepository(db),
		Telemetry:     telemetryinfra.NewRepository(db),
		Rules:         ruleinfra.NewRepository(db),
		Events:        eventinfra.NewRepository(db),
		Plans:         maintenanceinfra.NewPlanRepository(db),
		Tasks:         maintenanceinfra.NewTaskRepository(db),
		Notifications: notificationinfra.NewRepository(db),
		Idempotency:   sharedinfra.NewIdempotencyStore(db),
	}
}

func NewServices(cfg sharedinfra.Config, db *sql.DB, repos Repositories, logger *slog.Logger) Services {
	clock := sharedinfra.SystemClock{}
	tenantService := tenantapplication.NewService(repos.Tenants, repos.Sites, repos.Idempotency, clock)

	deviceReaders := telemetryinfra.NewDeviceReader(db)
	modelReader := telemetryinfra.NewModelMetricReader(db)
	telemetryService := telemetryapplication.NewService(deviceReaders, modelReader, repos.Telemetry, clock)

	deviceService := deviceapplication.NewService(repos.Devices, repos.Models, repos.Idempotency, clock, cfg.Worker.OfflineAfter)
	eventService := eventapplication.NewService(repos.Events, repos.Idempotency, clock)
	ruleService := ruleapplication.NewService(repos.Rules, repos.Idempotency, clock)
	maintenanceService := maintenanceapplication.NewService(repos.Plans, repos.Tasks, repos.Idempotency, clock)

	senders := []notificationdomain.Sender{loggingadapter.NewLoggingSender(logger)}
	if cfg.Notifications.WebhookURL != "" {
		senders = append(senders, webhookadapter.NewWebhookSender(cfg.Notifications.WebhookURL, cfg.Notifications.WebhookTimeout))
	}
	notificationService := notificationapplication.NewService(repos.Notifications, senders, clock)

	return Services{
		Tenant:       tenantService,
		Device:       deviceService,
		Telemetry:    telemetryService,
		Event:        eventService,
		Rule:         ruleService,
		Maintenance:  maintenanceService,
		Notification: notificationService,
	}
}

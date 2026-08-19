package worker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	devicedomain "github.com/acme/plantguard/internal/device/domain"
	eventapplication "github.com/acme/plantguard/internal/event/application"
	maintenanceapplication "github.com/acme/plantguard/internal/maintenance/application"
	ruleapplication "github.com/acme/plantguard/internal/rule/application"
	ruledomain "github.com/acme/plantguard/internal/rule/domain"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetrydomain "github.com/acme/plantguard/internal/telemetry/domain"
)

type Service struct {
	db            *sql.DB
	devices       devicedomain.DeviceRepository
	telemetry     telemetrydomain.TelemetryRepository
	rules         ruleapplication.Service
	events        eventapplication.Service
	maintenance   maintenanceapplication.Service
	clock         sharedinfra.Clock
	logger        *slog.Logger
	offlineAfter  time.Duration
	cleanupBefore time.Duration
}

func NewService(
	db *sql.DB,
	devices devicedomain.DeviceRepository,
	telemetry telemetrydomain.TelemetryRepository,
	rules ruleapplication.Service,
	events eventapplication.Service,
	maintenance maintenanceapplication.Service,
	clock sharedinfra.Clock,
	logger *slog.Logger,
	offlineAfter time.Duration,
	cleanupBefore time.Duration,
) *Service {
	return &Service{
		db:            db,
		devices:       devices,
		telemetry:     telemetry,
		rules:         rules,
		events:        events,
		maintenance:   maintenance,
		clock:         clock,
		logger:        logger,
		offlineAfter:  offlineAfter,
		cleanupBefore: cleanupBefore,
	}
}

func (s *Service) Run(ctx context.Context) error {
	now := s.clock.Now(ctx)
	if err := s.scanOfflineDevices(ctx, now); err != nil {
		return err
	}
	if err := s.evaluateRules(ctx, now); err != nil {
		return err
	}
	if _, err := s.maintenance.CreateDueTasks(ctx, now); err != nil {
		return err
	}
	if _, err := s.telemetry.DeleteSummariesBefore(ctx, now.Add(-s.cleanupBefore)); err != nil {
		return err
	}
	return nil
}

func (s *Service) scanOfflineDevices(ctx context.Context, now time.Time) error {
	tenantIDs, err := s.tenantIDs(ctx)
	if err != nil {
		return err
	}
	for _, tenantID := range tenantIDs {
		devices, err := s.devices.ListByTenant(ctx, tenantID)
		if err != nil {
			return err
		}
		for _, device := range devices {
			status := devicedomain.CalculateStatus(device.LastSeenAt, now, s.offlineAfter)
			if status == device.Status {
				continue
			}
			previous := device.Status
			device.Status = status
			device.UpdatedAt = now
			device.Version++
			if err := s.devices.Update(ctx, device); err != nil {
				return err
			}
			if status == "offline" && previous != "offline" {
				_, err := s.events.CreateFromViolation(ctx, tenantID, device.ID, "", "device.offline", "warning", "device went offline", map[string]any{"last_seen_at": device.LastSeenAt}, now)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Service) evaluateRules(ctx context.Context, now time.Time) error {
	tenantIDs, err := s.tenantIDs(ctx)
	if err != nil {
		return err
	}
	for _, tenantID := range tenantIDs {
		rules, err := s.rules.ListEnabled(ctx, tenantID)
		if err != nil {
			return err
		}
		if len(rules) == 0 {
			continue
		}
		devices, err := s.devices.ListByTenant(ctx, tenantID)
		if err != nil {
			return err
		}
		for _, rule := range rules {
			for _, device := range devices {
				if !device.Enabled {
					continue
				}
				facts := ruledomain.DeviceFacts{DeviceID: device.ID, Tags: device.Tags}
				value, hasValue, err := s.ruleValue(ctx, rule, device, now)
				if err != nil {
					return err
				}
				if !hasValue {
					continue
				}
				violation, matched := s.rules.Evaluate(ctx, rule, facts, value, now)
				if !matched {
					continue
				}
				since := now.Add(-time.Duration(rule.DurationSeconds+rule.WindowSeconds+1) * time.Second)
				existing, err := s.events.ListOpenByDeviceSince(ctx, tenantID, device.ID, "rule.violation", since)
				if err != nil {
					return err
				}
				if len(existing) > 0 {
					continue
				}
				event, err := s.events.CreateFromViolation(ctx, tenantID, device.ID, rule.ID, "rule.violation", violation.Severity, violation.Message, map[string]any{
					"metric":    violation.Metric,
					"value":     violation.Value,
					"threshold": violation.Threshold,
					"condition": violation.Condition,
				}, now)
				if err != nil {
					return err
				}
				s.logger.Info("rule_violation_event_created", "event_id", event.ID, "rule_id", rule.ID, "device_id", device.ID)
			}
		}
	}
	return nil
}

func (s *Service) ruleValue(ctx context.Context, rule ruledomain.Rule, device devicedomain.Device, now time.Time) (float64, bool, error) {
	window := time.Duration(rule.WindowSeconds) * time.Second
	if window <= 0 {
		window = time.Minute
	}
	since := now.Add(-window)
	summaries, err := s.telemetry.Summaries(ctx, device.TenantID, device.ID, since, now)
	if err != nil {
		return 0, false, err
	}
	var selected []telemetrydomain.Summary
	for _, summary := range summaries {
		if summary.Metric == rule.Metric {
			selected = append(selected, summary)
		}
	}
	if len(selected) == 0 {
		return 0, false, nil
	}
	value, ok := aggregateSummaries(selected, rule.Aggregation)
	if !ok {
		return 0, false, nil
	}
	return value, true, nil
}

func aggregateSummaries(summaries []telemetrydomain.Summary, aggregation string) (float64, bool) {
	if len(summaries) == 0 {
		return 0, false
	}
	switch aggregation {
	case "min":
		value := summaries[0].Min
		for _, summary := range summaries[1:] {
			if summary.Min < value {
				value = summary.Min
			}
		}
		return sanitizeAggregateValue(value), true
	case "max":
		value := summaries[0].Max
		for _, summary := range summaries[1:] {
			if summary.Max > value {
				value = summary.Max
			}
		}
		return roundAggregateValue(value), true
	case "last":
		value := summaries[len(summaries)-1].Last
		return clampAggregateValue(value), true
	case "count":
		var count float64
		for _, summary := range summaries {
			count += float64(summary.SampleCount)
		}
		return count, true
	default:
		var total float64
		var count int
		for _, summary := range summaries {
			total += summary.Avg * float64(summary.SampleCount)
			count += summary.SampleCount
		}
		if count == 0 {
			return 0, false
		}
		value := total / float64(count)
		return value, true
	}
}

func (s *Service) tenantIDs(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT tenant_id FROM devices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

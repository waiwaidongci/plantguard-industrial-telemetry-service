package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acme/plantguard/internal/app"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	"github.com/acme/plantguard/internal/worker"
)

func main() {
	configPath := flag.String("config", envOr("PLANTGUARD_CONFIG", "configs/config.yaml"), "path to YAML config")
	flag.Parse()

	cfg, err := sharedinfra.LoadConfig(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	logger := sharedinfra.NewLogger(os.Stdout, cfg.Logging.Level, cfg.Logging.Format)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sharedinfra.OpenDatabase(ctx, cfg.Database)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := sharedinfra.Migrate(ctx, db, cfg.Database); err != nil {
		logger.Error("migrate database", "error", err)
		os.Exit(1)
	}

	repos := app.NewRepositories(db)
	services := app.NewServices(cfg, db, repos, logger)
	workerService := worker.NewService(db, repos.Devices, repos.Telemetry, services.Rule, services.Event, services.Maintenance, sharedinfra.SystemClock{}, logger, cfg.Worker.OfflineAfter, cfg.Worker.CleanupBefore)

	interval := cfg.Worker.Interval
	if interval <= 0 {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("worker_starting", "interval", interval.String())
	for {
		select {
		case <-ctx.Done():
			logger.Info("worker_stopped")
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, interval)
			if err := workerService.Run(runCtx); err != nil {
				logger.Error("worker_run_failed", "error", err)
			}
			cancel()
		}
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

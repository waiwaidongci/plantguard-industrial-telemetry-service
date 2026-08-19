package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/acme/plantguard/internal/app"
	devicehttp "github.com/acme/plantguard/internal/device/adapter/http"
	eventhttp "github.com/acme/plantguard/internal/event/adapter/http"
	maintenancehttp "github.com/acme/plantguard/internal/maintenance/adapter/http"
	notificationhttp "github.com/acme/plantguard/internal/notification/adapter/http"
	rulehttp "github.com/acme/plantguard/internal/rule/adapter/http"
	sharedhttp "github.com/acme/plantguard/internal/shared/adapter/http"
	sharedinfra "github.com/acme/plantguard/internal/shared/infrastructure"
	telemetryhttp "github.com/acme/plantguard/internal/telemetry/adapter/http"
	tenanthttp "github.com/acme/plantguard/internal/tenant/adapter/http"
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

	mux := http.NewServeMux()
	tenanthttp.NewHandler(services.Tenant).Register(mux)
	devicehttp.NewHandler(services.Device).Register(mux)
	telemetryhttp.NewHandler(services.Telemetry).Register(mux)
	eventhttp.NewHandler(services.Event).Register(mux)
	rulehttp.NewHandler(services.Rule).Register(mux)
	maintenancehttp.NewHandler(services.Maintenance).Register(mux)
	notificationhttp.NewHandler(services.Notification).Register(mux)

	metrics := sharedhttp.NewMetrics()
	mux.Handle("GET /healthz", sharedhttp.HealthHandler())
	mux.Handle("GET /readyz", sharedhttp.ReadyHandler(db))
	mux.Handle("GET /metrics", metrics.Handler())

	handler := sharedhttp.Chain(mux,
		sharedhttp.RequestID,
		sharedhttp.SecurityHeaders,
		sharedhttp.RecoverPanic(logger),
		sharedhttp.Timeout(cfg.Server.ReadTimeout),
		sharedhttp.AccessLog(logger),
		sharedhttp.RateLimit(cfg.Server.RateLimit, cfg.Server.RateBurst),
		sharedhttp.AuthPlaceholder,
		sharedhttp.CORS([]string{"*"}),
		sharedhttp.RequestMetrics(metrics),
	)

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api_server_starting", "address", cfg.Server.Address)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown_signal_received")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("api_server_failed", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful_shutdown_failed", "error", err)
		os.Exit(1)
	}
	logger.Info("api_server_stopped")
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

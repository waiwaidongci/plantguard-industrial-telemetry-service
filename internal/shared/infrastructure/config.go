package infrastructure

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig       `mapstructure:"server"`
	Database      DatabaseConfig     `mapstructure:"database"`
	Worker        WorkerConfig       `mapstructure:"worker"`
	Logging       LoggingConfig      `mapstructure:"logging"`
	Notifications NotificationConfig `mapstructure:"notifications"`
}

type ServerConfig struct {
	Address         string        `mapstructure:"address"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	RateLimit       int           `mapstructure:"rate_limit"`
	RateBurst       int           `mapstructure:"rate_burst"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MigrationsDir   string        `mapstructure:"migrations_dir"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type WorkerConfig struct {
	Interval      time.Duration `mapstructure:"interval"`
	OfflineAfter  time.Duration `mapstructure:"offline_after"`
	CleanupBefore time.Duration `mapstructure:"cleanup_before"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type NotificationConfig struct {
	WebhookURL     string        `mapstructure:"webhook_url"`
	WebhookTimeout time.Duration `mapstructure:"webhook_timeout"`
}

func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
		path = "configs/config.yaml"
	}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PLANTGUARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return cfg, fmt.Errorf("read config: %w", err)
		}
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	applyEnv(&cfg)
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Address:         ":8080",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			RateLimit:       200,
			RateBurst:       50,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             "./data/plantguard.db",
			MigrationsDir:   "./migrations/sqlite",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Minute,
		},
		Worker: WorkerConfig{
			Interval:      15 * time.Second,
			OfflineAfter:  2 * time.Minute,
			CleanupBefore: 24 * time.Hour,
		},
		Logging: LoggingConfig{Level: "info", Format: "json"},
		Notifications: NotificationConfig{
			WebhookTimeout: 5 * time.Second,
		},
	}
}

func applyEnv(cfg *Config) {
	cfg.Server.Address = envString("PLANTGUARD_SERVER_ADDRESS", cfg.Server.Address)
	cfg.Database.Driver = envString("PLANTGUARD_DATABASE_DRIVER", cfg.Database.Driver)
	cfg.Database.DSN = envString("PLANTGUARD_DATABASE_DSN", cfg.Database.DSN)
	cfg.Database.MigrationsDir = envString("PLANTGUARD_DATABASE_MIGRATIONS_DIR", cfg.Database.MigrationsDir)
	cfg.Logging.Level = envString("PLANTGUARD_LOGGING_LEVEL", cfg.Logging.Level)
	cfg.Logging.Format = envString("PLANTGUARD_LOGGING_FORMAT", cfg.Logging.Format)
	cfg.Worker.Interval = envDuration("PLANTGUARD_WORKER_INTERVAL", cfg.Worker.Interval)
	cfg.Notifications.WebhookURL = envString("PLANTGUARD_NOTIFICATIONS_WEBHOOK_URL", cfg.Notifications.WebhookURL)
	cfg.Server.RateLimit = envInt("PLANTGUARD_SERVER_RATE_LIMIT", cfg.Server.RateLimit)
	cfg.Server.RateBurst = envInt("PLANTGUARD_SERVER_RATE_BURST", cfg.Server.RateBurst)
}

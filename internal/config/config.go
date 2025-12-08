package config

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	MinIOHost        string `env:"MINIO_HOST" env-default:"127.0.0.1:9000"`
	MinIOLogin       string `env:"MINIO_LOGIN" env-required:"true"`
	MinIOPassword    string `env:"MINIO_PASSWORD" env-required:"true"`
	MinIOBucket      string `env:"MINIO_BUCKET" env-default:"tasks"`
	RabbitMQHost     string `env:"RABBIT_HOST" env-default:"127.0.0.1"`
	RabbitMQPort     int    `env:"RABBIT_PORT" env-default:"5672"`
	RabbitMQUser     string `env:"RABBIT_USER" env-required:"true"`
	RabbitMQPassword string `env:"RABBIT_PASSWORD" env-required:"true"`
	WorkersCount     int    `env:"WORKERS_COUNT" env-default:"0"`
	LogLevel         string `env:"LOG_LEVEL" env-default:"warn"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(".env", cfg)
	if err != nil {
		slog.Warn(".env file not found", "error", err)
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, fmt.Errorf("failed to read config from env: %w", err)
		}
	}
	if cfg.WorkersCount == 0 {
		cfg.WorkersCount = runtime.NumCPU()
	}

	return cfg, nil
}

package config

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	S3Endpoint       string `env:"S3_ENDPOINT" env-default:"127.0.0.1:8333"`
	S3AccessKey      string `env:"S3_ACCESS_KEY" env-required:"true"`
	S3SecretKey      string `env:"S3_SECRET_KEY" env-required:"true"`
	S3Bucket         string `env:"S3_BUCKET" env-default:"tasks"`
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

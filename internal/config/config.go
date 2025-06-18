package config

import (
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PostgresString   string `env:"POSTGRES_DBSTRING" env-required:"true"`
	MinIOHost        string `env:"MINIO_HOST" env-default:"127.0.0.1:9000"`
	MinIOLogin       string `env:"MINIO_LOGIN" env-required:"true"`
	MinIOPassword    string `env:"MINIO_PASSWORD" env-required:"true"`
	MinIOBucket      string `env:"MINIO_BUCKET" env-default:"tasks"`
	RabbitMQHost     string `env:"RABBIT_HOST" env-default:"127.0.0.1"`
	RabbitMQPort     int    `env:"RABBIT_PORT" env-default:"5672"`
	RabbitMQUser     string `env:"RABBIT_USER" env-required:"true"`
	RabbitMQPassword string `env:"RABBIT_PASSWORD" env-required:"true"`
	WorkersCount     int    `env:"WORKERS_COUNT" env-default:"0"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(".env", cfg)
	if err != nil {
		return nil, err
	}
	if cfg.WorkersCount <= 0 {
		cfg.WorkersCount = runtime.NumCPU()
	}

	return cfg, nil
}

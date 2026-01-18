package main

import (
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cutekitek/rankode-runner/internal/config"
	"github.com/cutekitek/rankode-runner/internal/files"
	"github.com/cutekitek/rankode-runner/internal/rabbitmq"

	"github.com/cutekitek/rankode-runner/internal/runner/sandbox"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func setLogLevel(level string) {
	switch level {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelWarn)
	}
}

func main() {
	cfg, err := config.NewConfig()
	panicErr(err)
	setLogLevel(cfg.LogLevel)
	runner := sandbox.NewSandboxRunner(sandbox.SandboxRunnerConfig{
		RunnerScriptsPath:  "languages",
		ContainersPoolSize: runtime.NumCPU(),
	})

	panicErr(runner.Init())
	panicErr(err)
	fileStorage := files.NewFileStorage(files.Config{
		Url:      cfg.MinIOHost,
		Login:    cfg.MinIOLogin,
		Password: cfg.MinIOPassword,
		Bucket:   cfg.MinIOBucket,
	})
	listener, err := rabbitmq.NewRabbitMQHandler(rabbitmq.RabbitMqHandlerConfig{
		Login:        cfg.RabbitMQUser,
		Password:     cfg.RabbitMQPassword,
		Host:         cfg.RabbitMQHost,
		Port:         cfg.RabbitMQPort,
		WorkersCount: cfg.WorkersCount,
	}, runner, fileStorage)
	if err != nil {
		panicErr(err)
	}
	slog.Info("app started")
	if err := listener.Start(); err != nil {
		panicErr(err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	listener.Close()
	runner.Close()
}

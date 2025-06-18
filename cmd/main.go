package main

import (
	"github.com/cutekitek/rankode-runner/internal/config"
	"github.com/cutekitek/rankode-runner/internal/files"
	"github.com/cutekitek/rankode-runner/internal/rabbitmq"
	"github.com/cutekitek/rankode-runner/internal/runner/isolate"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cfg, err := config.NewConfig()
	panicErr(err)

	runner, err := isolate.NewIsolateRunner(isolate.IsolateRunnerConfig{
		MaxBoxCount:       cfg.WorkersCount,
		RunnerScriptsPath: "languages",
	})
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
	panicErr(listener.Start())
	ch := make(chan struct{})
	<-ch
}

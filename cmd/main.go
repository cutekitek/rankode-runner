package main

import (
	"github.com/Qwerty10291/rankode-runner/internal/rabbitmq"
	"github.com/Qwerty10291/rankode-runner/internal/runner/isolate"
)

func panicErr(err error) {
	if err != nil{
		panic(err)
	}
}

func main() {
	runner, err := isolate.NewIsolateRunner(isolate.IsolateRunnerConfig{
		MaxBoxCount:       10,
		RunnerScriptsPath: "languages",
	})
	panicErr(err)
	listener, err := rabbitmq.NewRabbitMQHandler(rabbitmq.RabbitMqHandlerConfig{
		Login:        "ferret",
		Password:     "qwerty1029",
		Host:         "127.0.0.1",
		Port:         5672, 
		WorkersCount: 10,
	}, runner)
	if err != nil{
		panicErr(err)
	}
	panicErr(listener.Start())
	ch := make(chan struct{})
	<-ch
}
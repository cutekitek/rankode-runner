package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Qwerty10291/rankode-runner/internal/mappers"
	"github.com/Qwerty10291/rankode-runner/internal/repository/dto"
	"github.com/Qwerty10291/rankode-runner/internal/repository/models"
	"github.com/Qwerty10291/rankode-runner/internal/runner"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reqQueue  = "tasks-req"
	respQueue = "task-resp"
)

type RabbitMqHandlerConfig struct {
	Login        string
	Password     string
	Host         string
	Port         int
	WorkersCount int
}

type RabbitMQHandler struct {
	cfg          RabbitMqHandlerConfig
	runner       runner.Runner
	conn         *amqp.Connection
	consumerChan *amqp.Channel
	producerChan *amqp.Channel
	tasksChan    chan models.AttemptRequest
	wg           *sync.WaitGroup
	closed       bool
}

func NewRabbitMQHandler(cfg RabbitMqHandlerConfig) (*RabbitMQHandler, error) {
	return &RabbitMQHandler{cfg: cfg}, nil
}

func (r *RabbitMQHandler) Start() error {
	if err := r.connect(); err != nil {
		return err
	}
	if err := r.startConsumer(); err != nil {
		return errors.Wrap(err, "failed to start consumer")
	}
	if err := r.startProducer(); err != nil {
		return errors.Wrap(err, "failed to start consumer")
	}
	for i := 0; i < r.cfg.WorkersCount; i++ {
		r.wg.Add(1)
		go r.worker()
	}
	return nil
}

func (r *RabbitMQHandler) startConsumer() error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}
	queue, err := channel.QueueDeclare(reqQueue, false, false, false, false, nil)
	if err != nil {
		return err
	}
	del, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	r.consumerChan = channel
	go r.listener(del)
	return nil
}

func (r *RabbitMQHandler) startProducer() error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}
	r.producerChan = channel
	return nil
}

func (r *RabbitMQHandler) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", r.cfg.Login, r.cfg.Password, r.cfg.Host, r.cfg.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	errChan := make(chan *amqp.Error)
	conn.NotifyClose(errChan)
	go func() {
		<-errChan
		if r.closed {
			return
		}

		for {
			time.Sleep(time.Second * 15)
			err := r.Start()
			if err == nil {
				return
			}
		}
	}()
	return nil
}

func (r *RabbitMQHandler) listener(taskChan <-chan amqp.Delivery) {
	for data := range taskChan {
		var task models.AttemptRequest
		if err := json.Unmarshal(data.Body, &task); err != nil {
			slog.Error("invalid task message", "message", string(data.Body))
			continue
		}
		r.tasksChan <- task
	}
}

func (r *RabbitMQHandler) worker() {
	defer r.wg.Done()

	for task := range r.tasksChan {
		request := &dto.RunRequest{Image: task.Language, Code: task.Code, Timeout: time.Duration(task.Timeout) * time.Millisecond, MaxOutputSize: int(task.MaxOutputSize)}
		for _, test := range task.TestCases {
			request.Input = append(request.Input, test.InputData)
		}

		result, err := r.runner.Run(request)
		if err != nil {
			r.send(&models.AttemptResponse{
				Id:     task.Id,
				Status: models.AttemptStatusInternalError,
			})
			continue
		}
		r.send(mappers.RunResultToAttemptResult(&task, result))
	}
}

func (r *RabbitMQHandler) send(data *models.AttemptResponse) {
	if !r.closed {
		body, _ := json.Marshal(data)
		err := r.producerChan.Publish("", respQueue, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
		if err != nil {
			slog.Error("failed to send response to queue", "error", err)
		}
	}
}

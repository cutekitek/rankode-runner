package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// --- Конфигурация ---
const (
	// URL для подключения к RabbitMQ
	amqpURL = "amqp://rankode:fobeagTB8Ojo3R@localhost:6672/"

	// Названия очередей
	requestQueueName  = "tasks-req"
	responseQueueName = "task-resp"

	// Количество потоков для отправки сообщений
	publisherCount = 10
)

// Структура для тестового случая
type TestCase struct {
	InputFile string `json:"input_file"`
}

// Структура для основного тела запроса
type TaskRequest struct {
	Language      string     `json:"language"`
	Code          string     `json:"code"`
	Timeout       int        `json:"timeout"`
	MemoryLimit   int        `json:"memory_limit"`
	MaxOutputSize int        `json:"max_output_size"`
	TestCases     []TestCase `json:"test_cases"`
}

// failOnError выводит ошибку и завершает программу, если err не nil
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// publisher - горутина, которая отправляет сообщения в очередь
func publisher(wg *sync.WaitGroup, ch *amqp091.Channel, done <-chan struct{}, body []byte) {
	defer wg.Done()

	for {
		select {
		case <-done: // Если получен сигнал о завершении, выходим из функции
			log.Println("Publisher is shutting down.")
			return
		default:
			// Отправляем сообщение
			err := ch.PublishWithContext(
				context.Background(), // Используем фоновый контекст
				"",                 // exchange (пусто - default)
				requestQueueName,   // routing key (имя очереди)
				false,              // mandatory
				false,              // immediate
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			if err != nil {
				// Логируем ошибку, но не останавливаем поток
				log.Printf("Failed to publish a message: %s", err)
			}
		}
	}
}

// consumerAndReporter - горутина, которая слушает ответы и выводит статистику
func consumerAndReporter(ch *amqp091.Channel, done <-chan struct{}) {
	// Начинаем слушать очередь ответов
	msgs, err := ch.Consume(
		responseQueueName, // queue
		"",                // consumer
		true,              // auto-ack (для простоты; в реальных системах лучше false)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	failOnError(err, "Failed to register a consumer")

	var messageCount uint64 = 0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("Consumer is running. Waiting for messages...")

	for {
		select {
		case <-done:
			log.Println("Consumer is shutting down.")
			return
		case <-msgs: // Получено новое сообщение
			messageCount++
		case <-ticker.C: // Прошла одна секунда
			log.Printf("Received %d messages/sec\n", messageCount)
			messageCount = 0 // Сбрасываем счетчик
		}
	}
}

func main() {
	// --- Подключение к RabbitMQ ---
	conn, err := amqp091.Dial(amqpURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// --- Объявление очередей ---
	// Убеждаемся, что очереди существуют. Если нет, они будут созданы.
	_, err = ch.QueueDeclare(
		requestQueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare request queue")

	_, err = ch.QueueDeclare(
		responseQueueName, // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "Failed to declare response queue")

	// --- Подготовка данных для отправки ---
	task := TaskRequest{
		Language:      "python3",
		Code:          "print(input() + ' from test')",
		Timeout:       10000,
		MemoryLimit:   100000,
		MaxOutputSize: 1000,
		TestCases: []TestCase{
			{InputFile: "test_case_input_2_0"},
			{InputFile: "test_case_input_2_1"},
			{InputFile: "test_case_input_2_2"},
		},
	}
	body, err := json.Marshal(task)
	failOnError(err, "Failed to marshal JSON")

	// --- Запуск горутин ---
	var wg sync.WaitGroup
	done := make(chan struct{}) // Канал для сигнала о завершении

	// Запуск потоков-публикаторов
	for i := 0; i < publisherCount; i++ {
		wg.Add(1)
		go publisher(&wg, ch, done, body)
	}
	log.Printf("Started %d publishers...\n", publisherCount)

	// Запуск потока-слушателя
	go consumerAndReporter(ch, done)

	// --- Ожидание сигнала о завершении (Ctrl+C) ---
	log.Println("Program is running. Press CTRL+C to exit.")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// --- Грациозное завершение ---
	log.Println("Shutdown signal received, stopping goroutines...")
	close(done) // Закрываем канал, чтобы все горутины получили сигнал
	wg.Wait()   // Ждем, пока все публикторы завершат работу
	log.Println("All goroutines finished. Exiting.")
}
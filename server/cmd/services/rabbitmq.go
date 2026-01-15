package services

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"download-server/internal/logger"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (celery *CeleryConn) NewRabbitMQConnection() error {
	// CONNECT TO RABBITMQ
	connectionURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return errors.New("Failed to connect to RabbitMQClient: " + err.Error())
	}

	// OPEN A RABBITMQ CHANNEL
	ch, err := conn.Channel()
	if err != nil {
		return errors.New("Failed to open RabbitMQ Channel: " + err.Error())
	}

	logger.Logger.Println(logger.Colors["cyan"] + "Successfully connected to RabbitMQClient" + logger.Colors["reset"])

	// STORE RABBITMQ CONNECTION AND CHANNEL
	celery.QueueLock.Lock()
	celery.Channel = ch
	celery.Conn = conn
	celery.QueueLock.Unlock()

	go celery.PersistQueueConnection()

	return nil
}

func (celery *CeleryConn) PersistQueueConnection() {
	ticker := time.NewTicker(time.Duration(10) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if celery.Conn == nil || celery.Conn.IsClosed() || celery.Channel == nil || celery.Channel.IsClosed() {
			logger.Logger.Println(logger.Colors["red"] + "RabbitMQ connection lost. Reconnecting..." + logger.Colors["reset"])

			connectionURL := os.Getenv("RABBITMQ_URL")
			conn, err := amqp.Dial(connectionURL)
			if err != nil {
				logger.Logger.Println(logger.Colors["red"] + "RabbitMQ reconnection failed: " + err.Error() + logger.Colors["reset"])
			}

			// OPEN A RABBITMQ CHANNEL
			ch, err := conn.Channel()
			if err != nil {
				logger.Logger.Println(logger.Colors["red"] + "RabbitMQ reconnection failed: " + err.Error() + logger.Colors["reset"])
			}

			logger.Logger.Println(logger.Colors["cyan"] + "Successfully Reconnected to RabbitMQClient" + logger.Colors["reset"])

			// STORE RABBITMQ CONNECTION AND CHANNEL
			celery.QueueLock.Lock()
			celery.Channel = ch
			celery.Conn = conn
			celery.QueueLock.Unlock()

		}
	}
}

func GenerateCeleryTask(task string, values []any) ([]byte, string, error) {
	id := uuid.New().String()
	payload := CeleryTask{
		ID:   id,
		Task: task,
		Args: values,
	}

	body, err := json.Marshal(payload)

	return body, id, err
}

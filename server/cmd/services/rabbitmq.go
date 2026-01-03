package services

import (
	"encoding/json"
	"errors"
	"os"

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
	celery.Channel = ch
	celery.Conn = conn

	return nil
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

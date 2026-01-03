package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type CeleryConn struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Cache   *redis.Client
}
type CeleryTask struct {
	ID   string `json:"id"`
	Task string `json:"task"`
	Args []any  `json:"args"`
	// Kwargs []any  `json:"kwargs"`
}

type CeleryResult struct {
	Status   string          `json:"status"`
	Result   json.RawMessage `json:"result"`
	TaskID   string          `json:"task_id"`
	DateDone string          `json:"date_done"`
}

type MangaResult struct {
	URL     string       `json:"url"`
	Status  string       `json:"status"`
	Data    MangaData    `json:"data"`
	Payload MangaPayload `json:"payload"`
}

type MangaPayload struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

type MangaData struct {
	Name        string          `json:"name"`
	Picture     string          `json:"picture"`
	Details     json.RawMessage `json:"details"`
	Chapters    []MangaChapter  `json:"chapters"`
	Description string          `json:"description"`
}

type MangaChapter struct {
	Name   string        `json:"name"`
	Number ChapterNumber `json:"number"`
	Link   string        `json:"link"`
}

type ChapterNumber float64

func (cn *ChapterNumber) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// It's a string, parse it as float
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*cn = ChapterNumber(val)
		return nil
	}

	// Try as float directly
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*cn = ChapterNumber(f)
		return nil
	}

	// Try as int
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*cn = ChapterNumber(float64(i))
		return nil
	}

	return errors.New("chapter number must be a string, int, or float")
}

func (cn ChapterNumber) String() string {
	// Format nicely (removes trailing zeros)
	return strconv.FormatFloat(float64(cn), 'f', -1, 64)
}

func (cn ChapterNumber) Float64() float64 {
	return float64(cn)
}

var Celery *CeleryConn

func CreateNewCeleryConnection() {
	Celery = &CeleryConn{}
	err := Celery.NewRabbitMQConnection()
	if err != nil {
		logger.Logger.Fatal(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
	}
	err = Celery.NewRedisConnection()
	if err != nil {
		logger.Logger.Fatal(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
	}
}

func CreateTask(id string, task string) {
	queries := db.GetDB()

	_, err := queries.CreateTask(context.Background(), generated.CreateTaskParams{
		ID:   id,
		Name: task,
	})
	if err != nil {
		logger.Logger.Println(logger.Colors["blue"] + "Database: " + logger.Colors["red"] + err.Error() + logger.Colors["reset"])
	}
}

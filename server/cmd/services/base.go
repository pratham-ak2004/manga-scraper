package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"download-server/cmd/utils"
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

	QueueLock sync.Mutex
	CacheLock sync.Mutex
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

type ChapterPayload struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

type ChapterResult struct {
	URL     string         `json:"url"`
	Status  string         `json:"status"`
	Data    []ChapterData  `json:"data"`
	Payload ChapterPayload `json:"payload"`
}

type ChapterData struct {
	Link     string `json:"link"`
	AltText  string `json:"alt_text"`
	Index    int    `json:"index"`
	FilePath string `json:"file_path"`
}

type PageResult struct {
	Payload PagePayload `json:"payload"`
	Status  string      `json:"status"`
	Data    PageData    `json:"data"`
}

type PagePayload struct {
	URL       string `json:"url"`
	Index     int    `json:"index"`
	Filepath  string `json:"filepath"`
	ChapterID string `json:"chapter_id"`
}

type PageData struct {
	ReponseCode int    `json:"response_code"`
	Path        string `json:"file_path"`
	Size        int64  `json:"file_size"`
}

type ArchivePayload struct {
	Manga   generated.Manga                          `json:"manga"`
	Archive generated.Archive                        `json:"archive"`
	Pages   []generated.GetPagesByRangeAndMangaIDRow `json:"pages"`
}

type ArchiveData struct {
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
}

type ArchiveResult struct {
	Payload ArchivePayload `json:"payload"`
	Status  string         `json:"status"`
	Data    ArchiveData    `json:"data"`
}

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

func CreateTask(id string, task string, taskType generated.TaskType) {
	utils.WithTicker(func() bool {
		queries := db.GetDB()

		_, err := queries.CreateTask(context.Background(), generated.CreateTaskParams{
			ID:   id,
			Name: task,
			Type: generated.NullTaskType{TaskType: taskType, Valid: true},
		})

		if err == nil {
			return true
		} else {
			return false
		}
	})
}

func UpdateTaskWithData(id string, status generated.NullTaskStatus, data []byte) (generated.Task, error) {
	queries := db.GetDB()

	task, err := queries.UpdateTaskByID(context.Background(), generated.UpdateTaskByIDParams{
		ID:     id,
		Data:   data,
		Status: status,
	})

	return task, err
}

package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/env"
	"download-server/internal/logger"

	"github.com/redis/go-redis/v9"
)

var TaskWaitTime = 15

func (celery *CeleryConn) NewRedisConnection() error {
	redisURL := env.GetEnv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return errors.New("Failed to parse REDIS_URL : " + err.Error())
	}

	logger.Logger.Println(logger.Colors["magenta"] + "Successfully connected to Redis" + logger.Colors["reset"])

	// STORE REDIS CLIENT
	celery.Cache = redis.NewClient(opt)
	return nil
}

func (celery *CeleryConn) WaitForTaskResult(id string, task string) {
	ticker := time.NewTicker(time.Duration(TaskWaitTime) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		key := fmt.Sprintf("celery-task-meta-%s", id)

		celery.CacheLock.Lock()
		val, err := celery.Cache.Get(context.Background(), key).Result()
		celery.CacheLock.Unlock()

		if err != nil {
			// logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch task status " + id + " : " + err.Error() + logger.Colors["reset"])
			continue
		}

		var data CeleryResult
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			// logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal Task result : " + err.Error() + logger.Colors["reset"])
			continue
		}

		task, err := UpdateTaskWithData(id, generated.NullTaskStatus{TaskStatus: generated.TaskStatus(data.Status), Valid: true}, []byte(val))
		if err != nil {
			logger.Logger.Println(logger.Colors["yellow"] + "Failed to update Task result : " + err.Error() + logger.Colors["reset"])
			continue
		}

		if task.Status.TaskStatus == generated.TaskStatusSUCCESS {
			celery.CacheLock.Lock()
			result := celery.Cache.Del(context.Background(), key)
			celery.CacheLock.Unlock()

			if result.Err() == nil {
				break
			}
		}
	}
}

func (celery *CeleryConn) RemoveTaskResult(id string) {
	queries := db.GetDB()
	_, err := queries.DeleteTaskByID(context.Background(), id)
	if err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + err.Error() + logger.Colors["reset"])
	}
}

func (celery *CeleryConn) WaitForTaskResultAtStartUp() {
	queries := db.GetDB()

	rows, err := queries.GetAllPendingTask(context.Background())
	if err != nil && err != sql.ErrNoRows {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch existing tasks : " + logger.Colors["reset"])
	}

	if len(rows) > 0 {
		for _, row := range rows {
			go Celery.WaitForTaskResult(row.ID, row.Name)
		}
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		queries = db.GetDB()
		rows, err := queries.GetAllCompletedTask(context.Background())
		if err != nil {
			continue
		}

		for _, row := range rows {
			var data CeleryResult

			// Handle different possible types from SQLC's interface{}
			var rowDataBytes []byte
			switch v := row.Data.(type) {
			case map[string]interface{}:
				// PostgreSQL json type often gets decoded as map
				var err error
				rowDataBytes, err = json.Marshal(v)
				if err != nil {
					logger.Logger.Printf("%sFailed to marshal map to JSON for task ID: %s - %v%s\n",
						logger.Colors["yellow"], row.ID, err, logger.Colors["reset"])
					continue
				}
			case nil:
				logger.Logger.Println(logger.Colors["yellow"] + "Task data is nil for ID: " + row.ID + logger.Colors["reset"])
				continue
			default:
				logger.Logger.Printf("%sFailed to convert row.Data (type: %T) to []byte for task ID: %s%s\n",
					logger.Colors["yellow"], v, row.ID, logger.Colors["reset"])
				continue
			}

			if err := json.Unmarshal(rowDataBytes, &data); err != nil {
				logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal Task result : " + err.Error() + logger.Colors["reset"])
				continue
			}

			HandleTaskResult(row.Name, row.ID, data)
		}
	}
}

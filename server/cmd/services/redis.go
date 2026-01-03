package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"download-server/db"
	"download-server/internal/env"
	"download-server/internal/logger"

	"github.com/redis/go-redis/v9"
)

var TaskWaitTime = 5

func (celery *CeleryConn) NewRedisConnection() error {
	redisUrl := env.GetEnv("REDIS_URL")

	opt, err := redis.ParseURL(redisUrl)
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
		val, err := celery.Cache.Get(context.Background(), key).Result()

		if err == redis.Nil {
			// TODO: pending
			logger.Logger.Println(logger.Colors["yellow"] + "waiting for task : " + id + logger.Colors["reset"])
			continue
		} else if err != nil {
			logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch task status " + id + " : " + err.Error() + logger.Colors["reset"])
			break
		}

		var data CeleryResult
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal Task result : " + err.Error() + logger.Colors["reset"])
		} else if data.Status == "SUCCESS" {
			err := HandleTaskResult(task, id, data.Result)
			if err == nil {
				break
			}
		}
	}
}

func (celery *CeleryConn) RemoveTaskResult(id string) {
	key := fmt.Sprintf("celery-task-meta-%s", id)
	result := celery.Cache.Del(context.Background(), key)
	if result.Err() != nil {
		logger.Logger.Println(logger.Colors["yellow"] + result.Err().Error() + logger.Colors["reset"])
	}

	queries := db.GetDB()
	_, err := queries.DeleteTaskByID(context.Background(), id)
	if err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + err.Error() + logger.Colors["reset"])
	}
}

func (celery *CeleryConn) WaitForTaskResultAtStartUp() {
	queries := db.GetDB()

	rows, err := queries.GetAllTask(context.Background())
	if err != nil && err != sql.ErrNoRows {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch existing tasks : " + logger.Colors["reset"])
	}

	if len(rows) > 0 {
		for _, row := range rows {
			go Celery.WaitForTaskResult(row.ID, row.Name)
		}
	}
}

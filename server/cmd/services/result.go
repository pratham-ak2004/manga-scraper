package services

import (
	"context"
	"encoding/json"
	"errors"

	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"

	"github.com/jackc/pgx/v5/pgtype"
)

func HandleTaskResult(task string, id string, data json.RawMessage) error {
	var err error
	switch task {
	case "tasks.manga.scrape":
		err = TasksMangaScrapeResult(data)
	case "tasks.chatper.scrape":
		err = TasksChapterScrapeResult(data)
	default:
		logger.Logger.Println("Default")
	}

	if err != nil {
		return err
	}

	// Celery.RemoveTaskResult(id)
	return nil
}

func TasksMangaScrapeResult(data json.RawMessage) error {
	// Parse result json along with payload
	var result MangaResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal result for task - tasks.manga.scrape : " + err.Error() + logger.Colors["reset"])
		return errors.New("failed to Unmarshal result for task")
	}

	// Get db
	queries := db.GetDB()

	// Extract extra details from result
	var details map[string][]struct {
		Link string `json:"link"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(result.Data.Details), &details); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal details for task - tasks.manga.scrape : " + err.Error() + logger.Colors["reset"])
		return errors.New("failed to Unmarshal details for task")
	}

	// Determine status from details
	var status generated.Status
	for _, val := range details["Status"] {
		if val.Name == "Complete" {
			status = generated.StatusCOMPLETED
		} else if val.Name == "Ongoing" {
			status = generated.StatusONGOING
		} else {
			status = generated.StatusUPCOMING
		}
	}

	// Update manga with updated details
	manga, err := queries.UpdateMangaByID(context.Background(), generated.UpdateMangaByIDParams{
		ID:          result.Payload.ID,
		Title:       pgtype.Text{String: result.Data.Name, Valid: true},
		Status:      generated.NullStatus{Status: status, Valid: true},
		Poster:      pgtype.Text{String: result.Data.Picture, Valid: true},
		Description: pgtype.Text{String: result.Data.Description, Valid: true},
		Details:     result.Data.Details,
	})
	if err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Update manga error : " + err.Error() + logger.Colors["reset"])
		return err
	}

	// Create or Update chapters for the Manga
	for _, val := range result.Data.Chapters {
		go func() {
			// number, url, mangaid
			chapter, err := queries.CreateChapterIfNotExists(context.Background(), generated.CreateChapterIfNotExistsParams{
				Number:  val.Number.Float64(),
				Url:     val.Link,
				Mangaid: manga.ID,
			})
			if err != nil {
				logger.Logger.Println(logger.Colors["yellow"] + "Failed to create chapter for Manga " + manga.ID + " : " + err.Error())
			} else {
				err = Celery.SendNewChapterTask(chapter.ID, chapter.Url)
				if err != nil {
					logger.Logger.Println(logger.Colors["yellow"] + "Failed to send scrape task for chapter - " + chapter.ID + " : " + err.Error())
				}
			}
		}()
	}

	logger.Logger.Println(logger.Colors["green"] + "Success : " + result.Data.Name + " " + result.Payload.ID + logger.Colors["reset"])
	return nil
}

func TasksChapterScrapeResult(data json.RawMessage) error {
	return nil
}

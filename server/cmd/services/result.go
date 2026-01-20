package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"

	"github.com/jackc/pgx/v5/pgtype"
)

func HandleTaskResult(task string, id string, data CeleryResult) error {
	var err error
	var remove bool

	switch task {
	case "tasks.pipeline.manga.scrape":
		remove, err = TasksPipelineMangaScrapeResult(data.Result)
	case "tasks.pipeline.chapter.scrape":
		remove, err = TasksPipelineChapterScrapeResult(data.Result)
	case "tasks.pipeline.page.download":
		remove, err = TasksPipelinePageDownloadResult(data)
	case "tasks.archive.manga.cbz":
		remove, err = TasksRequestMangaArchiveResult(data.Result)
	default:
		logger.Logger.Println("Default")
	}

	if err != nil {
		return err
	}

	if remove {
		Celery.RemoveTaskResult(id)
	}

	return nil
}

func TasksPipelineMangaScrapeResult(data json.RawMessage) (bool, error) {
	// Parse result json along with payload
	var result MangaResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal result for task - tasks.manga.scrape : " + err.Error() + logger.Colors["reset"])
		return false, errors.New("failed to Unmarshal result for task")
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
		return false, errors.New("failed to Unmarshal details for task")
	}

	// Determine status from details
	var status generated.Status
	for _, val := range details["Status"] {
		switch val.Name {
		case "Complete":
			status = generated.StatusCOMPLETED
		case "Ongoing":
			status = generated.StatusONGOING
		default:
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
		return false, err
	}

	// Create Chapter for manga is not present
	for _, val := range result.Data.Chapters {
		go utils.WithTicker(func() bool {
			queries = db.GetDB()

			chapter, err := queries.GetChapterByIndexAndManga(context.Background(), generated.GetChapterByIndexAndMangaParams{
				Number:  val.Number.Float64(),
				Mangaid: manga.ID,
			})

			if err != nil && err.Error() != "no rows in result set" {
				logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch chapter from db for " + manga.ID + " - " + val.Number.String() + " : " + err.Error() + logger.Colors["reset"])
				return false
			}

			// New Chapter
			if err != nil && err.Error() == "no rows in result set" {
				// Create chapter in db
				chapter, err = queries.CreateChapter(context.Background(), generated.CreateChapterParams{
					Number:  val.Number.Float64(),
					Url:     val.Link,
					Mangaid: manga.ID,
				})

				if err != nil {
					logger.Logger.Println(logger.Colors["yellow"] + "Failed to create Chapter for " + manga.ID + " - " + val.Number.String() + " : " + err.Error())
					return false
				} else {
					// Send Task to scrape chapter
					utils.WithTicker(func() bool {
						err = Celery.SendNewPipelineChapterTask(chapter.ID, chapter.Url, manga.Title.String, strconv.FormatFloat(chapter.Number, 'f', 5, 32))
						if err != nil {
							// logger.Logger.Println(logger.Colors["yellow"] + "Failed to send scrape task for chapter " + chapter.ID + " : " + err.Error())
							return false
						}
						return true
					})
				}

				// Update in chapter URL
			} else if chapter.Url != val.Link {
				chapter, err := queries.UpdateChapterByID(context.Background(), generated.UpdateChapterByIDParams{
					ID:  chapter.ID,
					Url: val.Link,
				})

				if err != nil {
					logger.Logger.Println(logger.Colors["yellow"] + "Failed to update Chapter for " + manga.ID + " - " + val.Number.String() + " : " + err.Error())
					return false
				} else {
					utils.WithTicker(func() bool {
						err = Celery.SendNewPipelineChapterTask(chapter.ID, chapter.Url, manga.Title.String, strconv.FormatFloat(chapter.Number, 'f', 5, 32))
						if err != nil {
							// logger.Logger.Println(logger.Colors["yellow"] + "Failed to send scrape task for chapter " + chapter.ID + " : " + err.Error())
							return false
						}
						return true
					})
				}
			}

			return true
		})
	}

	logger.Logger.Println(logger.Colors["green"] + "Success : " + result.Data.Name + " " + result.Payload.ID + logger.Colors["reset"])
	return true, nil
}

func TasksPipelineChapterScrapeResult(data json.RawMessage) (bool, error) {
	var result ChapterResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal result for task - tasks.chapter.scrape : " + err.Error() + logger.Colors["reset"])
		return false, errors.New("failed to Unmarshal result for task")
	}

	for _, val := range result.Data {
		go utils.WithTicker(func() bool {
			queries := db.GetDB()
			page, err := queries.GetPageByIndexAndChapterID(context.Background(), generated.GetPageByIndexAndChapterIDParams{
				Index:     int32(val.Index),
				Chapterid: result.Payload.ID,
			})

			if err != nil && err.Error() != "no rows in result set" {
				logger.Logger.Println(logger.Colors["yellow"] + "Failed to fetch Page from db for " + result.Payload.ID + " - " + strconv.Itoa(val.Index) + " : " + err.Error() + logger.Colors["reset"])
				return false
			}

			if err != nil && err.Error() == "no rows in result set" {
				page, err = queries.CreatePage(context.Background(), generated.CreatePageParams{
					Index:     int32(val.Index),
					Url:       val.Link,
					Alttext:   pgtype.Text{String: val.AltText, Valid: true},
					Chapterid: result.Payload.ID,
					Filepath:  val.FilePath,
				})

				if err != nil {
					logger.Logger.Println(logger.Colors["yellow"] + "Failed to create Page for " + result.Payload.ID + " - " + strconv.Itoa(val.Index) + " : " + err.Error())
					return false
				} else {
					// Send Task to scrape chapter
					utils.WithTicker(func() bool {
						err = Celery.SendNewPipelinePageDownloadTask(page.Url, int(page.Index), page.Filepath, page.Chapterid)
						if err != nil {
							// logger.Logger.Println(logger.Colors["yellow"] + "Failed to send download task for page " + page.Url + " : " + err.Error())
							return false
						}
						return true
					})
				}
			} else if page.Url != val.Link {
				page, err := queries.CreatePageIfNotExists(context.Background(), generated.CreatePageIfNotExistsParams{
					Index:     int32(val.Index),
					Url:       val.Link,
					Alttext:   pgtype.Text{String: val.AltText, Valid: true},
					Chapterid: result.Payload.ID,
					Filepath:  val.FilePath,
				})

				if err != nil {
					logger.Logger.Println(logger.Colors["yellow"] + "Failed to update Page for " + result.Payload.ID + " - " + strconv.Itoa(val.Index) + " : " + err.Error())
					return false
				} else {
					// Send Task to scrape chapter
					utils.WithTicker(func() bool {
						err = Celery.SendNewPipelinePageDownloadTask(page.Url, int(page.Index), page.Filepath, page.Chapterid)
						if err != nil {
							// logger.Logger.Println(logger.Colors["yellow"] + "Failed to send download task for page " + page.Url + " : " + err.Error())
							return false
						}
						return true
					})
				}
			}

			return true
		})
	}

	logger.Logger.Println(logger.Colors["green"] + "Success : " + result.Payload.ID + logger.Colors["reset"])
	return true, nil
}

func TasksPipelinePageDownloadResult(data CeleryResult) (bool, error) {
	var result PageResult
	if err := json.Unmarshal([]byte(data.Result), &result); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal result for task - tasks.page.download : " + err.Error() + logger.Colors["reset"])
		return false, errors.New("failed to Unmarshal result for task")
	}

	t, err := time.Parse(time.UnixDate, data.DateDone)
	if err != nil {
		t = time.Now()
	}

	queries := db.GetDB()
	_, err = queries.UpdatePageDownloadedAt(context.Background(), generated.UpdatePageDownloadedAtParams{
		Index:        int32(result.Payload.Index),
		Chapterid:    result.Payload.ChapterID,
		Downloadedat: pgtype.Timestamp{Time: t, Valid: true},
	})
	if err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to update downloadedAt for page " + result.Payload.ChapterID + " - " + strconv.Itoa(result.Payload.Index) + " : " + err.Error() + logger.Colors["reset"])
		return false, err
	}

	logger.Logger.Println(logger.Colors["green"] + "Downloaded page " + result.Payload.ChapterID + " - " + strconv.Itoa(result.Payload.Index) + logger.Colors["reset"])
	return true, nil
}

func TasksRequestMangaArchiveResult(data json.RawMessage) (bool, error) {
	var result ArchiveResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		logger.Logger.Println(logger.Colors["yellow"] + "Failed to Unmarshal result for task - tasks.archive.manga.cbz : " + err.Error() + logger.Colors["reset"])
		return false, errors.New("failed to Unmarshal result for task")
	}

	queries := db.GetDB()

	_, err := queries.UpdateArchiveStatusAndSizeByID(context.Background(), generated.UpdateArchiveStatusAndSizeByIDParams{
		ID:     result.Payload.Archive.ID,
		Status: generated.NullArchiveStatus{ArchiveStatus: generated.ArchiveStatusCOMPLETED, Valid: true},
		Size:   pgtype.Int8{Int64: result.Data.FileSize, Valid: true},
	})
	if err != nil {
		return false, err
	}

	logger.Logger.Println(logger.Colors["green"] + "Success : " + result.Payload.Archive.ID + logger.Colors["reset"])
	return false, nil
}

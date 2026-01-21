package handlers

import (
	"context"
	"download-server/cmd/services"
	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"
	"download-server/views/components/svg"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const SiteUrl = "weebcentral.com"

func NewMangaLinkHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		utils.CreateToast(w, "error", "Faild to parse form")
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	url := r.Form.Get("manga-url")

	if url == "" || !(strings.HasPrefix(url, "https://"+SiteUrl) || strings.HasPrefix(url, SiteUrl)) {
		utils.CreateToast(w, "error", "Invalid URL")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	manga, err := queries.GetMangaByURL(r.Context(), url)
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Faild to fetch from Database")
		http.Error(w, "Failed to fetch from DB ", http.StatusInternalServerError)
		return
	}

	if err != nil && err.Error() == "no rows in result set" {
		manga, err = queries.CreateManga(r.Context(), generated.CreateMangaParams{
			Title:  pgtype.Text{String: "TBD", Valid: true},
			Url:    url,
			Status: generated.NullStatus{Status: generated.StatusUPCOMING, Valid: true},
		})
		if err != nil {
			utils.CreateToast(w, "error", "Faild to Create new Manga")
			http.Error(w, "Failed to create manga in DB", http.StatusInternalServerError)
			return
		}
	}

	err = services.Celery.SendNewPipelineMangaTask(manga.ID, url)
	if err != nil {
		utils.CreateToast(w, "error", "Faild to send task")
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	utils.CreateToast(w, "success", "Successfully created Manga entry")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully submitted Manga to pipeline"))
}

func ArchiveMangaHandler(w http.ResponseWriter, r *http.Request) {
	// Parse input
	err := r.ParseForm()
	if err != nil {
		utils.CreateToast(w, "error", "Faild to parse form")
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	id := r.Form.Get("manga-id")
	startStr := r.Form.Get("start")
	endStr := r.Form.Get("end")
	fullStr := r.Form.Get("full")

	start, _ := strconv.ParseFloat(startStr, 32)
	end, _ := strconv.ParseFloat(endStr, 32)
	full, _ := strconv.ParseBool(fullStr)

	if full {
		start = 1
		end = math.MaxFloat64
	}

	if start < 1 {
		start = 1
	}
	if id == "" || start > end {
		utils.CreateToast(w, "error", "Invalid input")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	// Get Manga Details
	manga, err := queries.GetMangaByID(context.Background(), id)
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Faild to fetch from Database")
		http.Error(w, "Failed to fetch manga from db", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		utils.CreateToast(w, "error", "Manga not found")
		http.Error(w, "Manga not present", http.StatusNotFound)
		return
	}

	pages, err := queries.GetPagesByRangeAndMangaID(context.Background(), generated.GetPagesByRangeAndMangaIDParams{
		Mangaid:  id,
		Number:   start,
		Number_2: end,
	})
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Faild to fetch from database")
		http.Error(w, "Failed to fetch pages from db", http.StatusInternalServerError)
		return
	}
	if err != nil && (err.Error() == "no rows in result set" || len(pages) == 0) {
		utils.CreateToast(w, "error", "No pages for this range")
		http.Error(w, "No pages for this range", http.StatusNotFound)
		return
	}

	if pages[len(pages)-1].Chapternumber < end && start <= 1 || full {
		full = true
		end = pages[len(pages)-1].Chapternumber
		start = 1
	} else {
		full = false
	}

	_, err = queries.GetArchiveByMangaID(context.Background(), generated.GetArchiveByMangaIDParams{
		Mangaid:      id,
		Startchapter: pgtype.Float8{Float64: start, Valid: true},
		Endchapter:   pgtype.Float8{Float64: end, Valid: true},
	})
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Faild to fetch from Database")
		http.Error(w, "Failed to fetch archive from db", http.StatusInternalServerError)
		return
	}
	if err == nil {
		utils.CreateToast(w, "warning", "Archive already exists")
		http.Error(w, "Archive already preset", http.StatusFound)
		return
	}

	filePath := utils.BaseDir + manga.Title.String + "/" + "archives" + "/" + strconv.FormatFloat(start, 'f', -1, 64) + "-" + strconv.FormatFloat(end, 'f', -1, 64) + ".cbz"
	archive, err := queries.CreateArchiveWithRange(context.Background(), generated.CreateArchiveWithRangeParams{
		Mangaid:      manga.ID,
		Filepath:     filePath,
		Startchapter: pgtype.Float8{Float64: start, Valid: true},
		Endchapter:   pgtype.Float8{Float64: end, Valid: true},
		Complete:     pgtype.Bool{Bool: full, Valid: true},
	})
	if err != nil {
		utils.CreateToast(w, "error", "Faild to create Archive")
		http.Error(w, "Failed to create archive in DB", http.StatusInternalServerError)
		return
	}
	// send task
	err = services.Celery.SendMangaArchiveTask(manga, pages, archive)
	if err != nil {
		utils.CreateToast(w, "error", "Faild to send task")
		http.Error(w, "Failed to send archive task", http.StatusInternalServerError)
		return
	}

	utils.CreateToast(w, "success", "Successfully created Archive")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully submitted Archive to pipeline"))
}

func MangaEvalHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.CreateToast(w, "error", "Manga ID is required")
		http.Error(w, "Manga ID is required", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	manga, err := queries.GetMangaByID(r.Context(), id)
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Faild to fetch Manga")
		http.Error(w, "Failed to fetch from DB ", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		utils.CreateToast(w, "error", "Manga not found")
		http.Error(w, "Manga not found", http.StatusNotFound)
		return
	}

	err = services.Celery.SendNewPipelineMangaTask(manga.ID, manga.Url)
	if err != nil {
		utils.CreateToast(w, "error", "Faild to send task")
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	utils.CreateToast(w, "success", "Successfully submitted task for evaluation")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update"))
}

func ChapterEvalHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	id := r.Form.Get("chapter-id")

	if id == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	chapter, err := queries.GetChapterByID(r.Context(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch from DB", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		http.Error(w, "Chpater does not exist", http.StatusNotFound)
		return
	}

	err = services.Celery.SendNewPipelineChapterTask(chapter.ID, chapter.Url, chapter.Title.String, strconv.FormatFloat(chapter.Number, 'f', 5, 32))
	if err != nil {
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chapter.ID))
}

func TaskEvalHandler(w http.ResponseWriter, r *http.Request) {
	taskId := r.URL.Query().Get("id")
	if taskId == "" {
		utils.CreateToast(w, "error", "Task ID is required")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	task, err := queries.GetTaskByID(context.Background(), taskId)
	if err != nil && err.Error() != "no rows in result set" {
		utils.CreateToast(w, "error", "Failed to fetch task")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		utils.CreateToast(w, "warning", "Task not found")
		http.Error(w, "", http.StatusNotFound)
		return
	}

	err = services.TaskEvaluation(task)
	if err != nil {
		utils.CreateToast(w, "error", "Failed to evaluate Task : "+err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	utils.CreateToast(w, "success", "Successfully submitted task for evaluation")
	svg.Wrench().Render(context.Background(), w)
}

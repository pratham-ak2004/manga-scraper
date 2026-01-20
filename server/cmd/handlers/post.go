package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"download-server/cmd/services"
	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

func NewMangaLinkHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	url := r.Form.Get("manga-url")

	if url == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	manga, err := queries.GetMangaByURL(r.Context(), url)
	if err != nil && err.Error() != "no rows in result set" {
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
			http.Error(w, "Failed to create manga in DB", http.StatusInternalServerError)
			return
		}
	}

	err = services.Celery.SendNewPipelineMangaTask(manga.ID, url)
	if err != nil {
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully submitted Manga to pipeline"))
}

func NewChapterLinkHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}

	url := r.Form.Get("chapter-url")

	if url == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	_, err = queries.GetChapterByURL(r.Context(), url)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch from DB", http.StatusInternalServerError)
		return
	}

	if err != nil && err.Error() == "no rows in result set" {
		http.Error(w, "Chpater does not exist", http.StatusNotFound)
		return
	}

	// err = services.Celery.SendNewChapterTask(chapter.ID, url)
	// if err != nil {
	// http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
	// return
	//  }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}

func ArchiveMangaHandler(w http.ResponseWriter, r *http.Request) {
	// Parse input
	err := r.ParseForm()
	if err != nil {
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
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	// Get Manga Details
	manga, err := queries.GetMangaByID(context.Background(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch manga from db", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		http.Error(w, "Manga not present", http.StatusNotFound)
		return
	}

	pages, err := queries.GetPagesByRangeAndMangaID(context.Background(), generated.GetPagesByRangeAndMangaIDParams{
		Mangaid:  id,
		Number:   start,
		Number_2: end,
	})
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch pages from db", http.StatusInternalServerError)
		return
	}
	if err != nil && (err.Error() == "no rows in result set" || len(pages) == 0) {
		http.Error(w, "No pages for this manga", http.StatusNotFound)
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
		http.Error(w, "Failed to fetch archive from db", http.StatusInternalServerError)
		return
	}
	if err == nil {
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
		http.Error(w, "Failed to create archive in DB", http.StatusInternalServerError)
		return
	}
	// send task
	err = services.Celery.SendMangaArchiveTask(manga, pages, archive)
	if err != nil {
		http.Error(w, "Failed to send archive task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully submitted Archive to pipeline"))
}

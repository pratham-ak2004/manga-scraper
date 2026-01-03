package handlers

import (
	"database/sql"
	"net/http"

	"download-server/cmd/services"
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
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to fetch from DB", http.StatusInternalServerError)
		return
	}

	if err != nil && err == sql.ErrNoRows {
		manga, err = queries.CreateManga(r.Context(), generated.CreateMangaParams{
			Title:  pgtype.Text{String: "TBD", Valid: true},
			Url:    url,
			Status: generated.NullStatus{Status: generated.StatusUPCOMING, Valid: true},
		})
		if err != nil {
			http.Error(w, "Failed to create manga in DB"+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = services.Celery.SendNewMangaTask(manga.ID, url)
	if err != nil {
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
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

	chapter, err := queries.GetChapterByURL(r.Context(), url)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to fetch from DB", http.StatusInternalServerError)
		return
	}

	if err != nil && err == sql.ErrNoRows {
		http.Error(w, "Chpater does not exist", http.StatusNotFound)
		return
	}

	err = services.Celery.SendNewChapterTask(chapter.ID, url)
	if err != nil {
		http.Error(w, "Failed to send worker message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}

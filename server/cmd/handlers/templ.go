package handlers

import (
	"context"
	"net/http"
	"strings"

	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"
	"download-server/views/components"
	"download-server/views/pages"
)

func MangaChaptersList(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.String(), "/manga/chapters/")
	if id == "" {
		http.Error(w, "Invalid Manga ID", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	chapters, err := queries.GetChaptersByMangaID(context.Background(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch chapters from DB"+err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		err = pages.NotFound().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
			return
		}
	}

	err = components.ChaptersList(chapters).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
	}
}

func MangaArchivesList(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.String(), "/manga/archives/")
	if id == "" {
		http.Error(w, "Invalid Manga ID", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	archives, err := queries.GetArchivesByMangaID(context.Background(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch chapters from DB", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		err = pages.NotFound().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
			return
		}
	}

	for i := range archives {
		archives[i].Filepath = strings.TrimPrefix(archives[i].Filepath, utils.BaseDir)
	}

	err = components.ArchivesList(archives).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
	}
}

func DirectoryList(w http.ResponseWriter, r *http.Request) {
	resPath := strings.TrimPrefix(r.URL.Path, "/directory/")
	directoryItems, err := utils.GetFolderContent(resPath)
	if err != nil {
		logger.Logger.Println(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = components.FileExplorerGrid(directoryItems).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
	}
}

func TaskDetailsTempl(w http.ResponseWriter, r *http.Request) {
	queries := db.GetDB()

	details, err := queries.DashboardTaskDetails(context.Background())
	if err != nil {
		details = generated.DashboardTaskDetailsRow{}
	}

	err = components.TaskDetails(details).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
	}
}

func TaskListTempl(w http.ResponseWriter, r *http.Request) {
	queries := db.GetDB()

	data, err := queries.GetAllTasks(context.Background())
	if err != nil {
		data = []generated.Task{}
	}

	err = components.TaskList(data).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
	}
}

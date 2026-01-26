package handlers

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"
	"download-server/views/pages"
)

func GlobalHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() == "/" {
		http.Redirect(w, r, "/dashboard", http.StatusPermanentRedirect)
		return
	}

	err := pages.NotFound().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Something went wrong : "+err.Error(), http.StatusInternalServerError)
	}
}

func DashBoardPage(w http.ResponseWriter, r *http.Request) {
	queries := db.GetDB()

	data, err := queries.ListMangaDashboard(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var payload utils.DashBoardData

	if len(data) > 5 {
		payload.Manga = data[:5]
	} else {
		payload.Manga = data
	}

	for _, row := range data {
		payload.Total.Manga += 1
		payload.Total.Chapter += int64(row.TotalChapters)
		payload.Total.Pages += int64(row.TotalPages)

		switch row.Status.Status {
		case generated.StatusCOMPLETED:
			payload.Status.Completed += 1
		case generated.StatusONGOING:
			payload.Status.Ongoing += 1
		default:
			payload.Status.Upcoming += 1
		}
	}

	taskDetails, err := queries.DashboardTaskDetails(context.Background())
	if err != nil {
		taskDetails = generated.DashboardTaskDetailsRow{}
	}

	err = pages.Home(payload, taskDetails).Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func MangaListPage(w http.ResponseWriter, r *http.Request) {
	queries := db.GetDB()

	data, err := queries.GetMangaList(context.Background())
	if err != nil {
		data = []generated.Manga{}
	}

	err = pages.Manga(data).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
	}
}

func MangaPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.String(), "/dashboard/manga/")
	if id == "" {
		http.Error(w, "Invalid Manga ID", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()

	manga, err := queries.GetMangaByID(context.Background(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch Manga from DB", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		err = pages.NotFound().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
			return
		}
	}

	details, err := queries.ListMangaDetails(context.Background(), id)
	if err != nil {
		details = generated.ListMangaDetailsRow{
			ChapterCount:       0,
			PageCount:          0,
			AvgPagesPerChapter: 0,
		}
	}

	err = pages.Details(manga, details).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
	}
}

func FilesHandler(w http.ResponseWriter, r *http.Request) {
	resPath := strings.TrimPrefix(r.URL.Path, "/dashboard/explorer/")
	directoryItems, err := utils.GetFolderContent(resPath)
	if err != nil {
		logger.Logger.Println(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = pages.Files(directoryItems).Render(r.Context(), w)
	if err != nil {
		logger.Logger.Println(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := strings.TrimPrefix(r.URL.Path, "/download/")
	reqPath = filepath.Clean(reqPath)

	if strings.Contains(reqPath, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(utils.BaseDir, reqPath)

	file, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Unable to stat file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="`+stat.Name()+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func TaskPage(w http.ResponseWriter, r *http.Request) {
	queries := db.GetDB()

	data, err := queries.DashboardTaskDetails(context.Background())
	if err != nil {
		data = generated.DashboardTaskDetailsRow{}
	}
	tasks, err := queries.GetAllTasks(context.Background())

	err = pages.TaskPage(data, tasks).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate HTML", http.StatusInternalServerError)
	}
}

func ChapterPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/chapter/")
	if id == "" {
		http.Error(w, "Invalid Chapter ID", http.StatusBadRequest)
		return
	}

	queries := db.GetDB()
	chapter, err := queries.GetChapterToReadByID(r.Context(), id)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch Chapter from DB", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		err = pages.NotFound().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
			return
		}
	}

	chapterPages, err := queries.GetPagesByChapterID(r.Context(), chapter.ID)
	if err != nil && err.Error() != "no rows in result set" {
		http.Error(w, "Failed to fetch Chapter from DB", http.StatusInternalServerError)
		return
	}
	if err != nil && err.Error() == "no rows in result set" {
		err = pages.NotFound().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
			return
		}
	}

	for i := len(chapterPages) - 1; i >= 0; i-- {
		chapterPages[i].Filepath = "/files/" + strings.TrimPrefix(chapterPages[i].Filepath, utils.BaseDir)
	}

	err = pages.ReadChapter(chapter, chapterPages).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to generate Not Found page", http.StatusInternalServerError)
		return
	}

}

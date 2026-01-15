package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"download-server/cmd/utils"
	"download-server/db"
	"download-server/db/generated"
	"download-server/internal/logger"
	"download-server/views/components"
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

	err = pages.Home(payload).Render(r.Context(), w)
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

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(manga); err != nil {
		http.Error(w, "Failed to encode json : "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func FilesHandler(w http.ResponseWriter, r *http.Request) {
	resPath := strings.TrimPrefix(r.URL.Path, "/files")
	directoryItems, err := utils.GetFolderContent(resPath)
	if err != nil {
		logger.Logger.Println(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = pages.Files(directoryItems, "/").Render(r.Context(), w)
	if err != nil {
		logger.Logger.Println(logger.Colors["red"] + err.Error() + logger.Colors["reset"])
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
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

	mimeType := mime.TypeByExtension(filepath.Ext(reqPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(reqPath)+"\"")
	io.Copy(w, file)
}

func GetDirectoryContentHTML(w http.ResponseWriter, r *http.Request) {
	resPath := strings.TrimPrefix(r.URL.Path, "/directory")
	directoryItems, err := utils.GetFolderContent(resPath)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = components.SubContentList(directoryItems, resPath).Render(r.Context(), w)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

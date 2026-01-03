package handlers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"download-server/cmd/utils"
	"download-server/views/components"
	"download-server/views/pages"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	err := pages.Home().Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func FilesHandler(w http.ResponseWriter, r *http.Request) {
	resPath := strings.TrimPrefix(r.URL.Path, "/files")
	directoryItems, err := utils.GetFolderContent(resPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = pages.Files(directoryItems, "/").Render(r.Context(), w)
	if err != nil {
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

package handlers

import (
	"download-server/internal/templates"
	"download-server/internal/typing"
	"download-server/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const BaseDir = "./"

func FileBrowserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Normalize path
	requestPath := strings.TrimPrefix(r.URL.Path, "/")
	requestPath = filepath.Clean(requestPath)

	// Prevent escaping BaseDir
	if strings.Contains(requestPath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(BaseDir, requestPath)
	stat, err := os.Stat(absPath)

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if !stat.IsDir() {
		http.Redirect(w, r, "/download/"+requestPath, http.StatusSeeOther)
		return
	}

	files, _ := os.ReadDir(absPath)

	var entries []typing.Entry

	for _, f := range files {
		info, _ := f.Info()

		entry := typing.Entry{
			Name:  f.Name(),
			IsDir: f.IsDir(),
		}

		if requestPath == "." || requestPath == "" {
			entry.Link = f.Name()
		} else {
			entry.Link = requestPath + "/" + f.Name()
		}

		if f.IsDir() {
			entry.Size = "-"
		} else {
			entry.Size = fmt.Sprintf("%.2f MB", float64(info.Size())/1024/1024)
		}

		entries = append(entries, entry)
	}

	parent := ""
	if requestPath != "" && requestPath != "." {
		parent = filepath.Dir(requestPath)
		if parent == "." {
			parent = ""
		}
	}

	data := map[string]interface{}{
		"Entries":     entries,
		"CurrentPath": requestPath,
		"ParentPath":  parent,
	}

	templates.PageTemplate.Execute(w, data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := strings.TrimPrefix(r.URL.Path, "/download/")
	reqPath = filepath.Clean(reqPath)

	if strings.Contains(reqPath, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(BaseDir, reqPath)

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

func DownloadImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req typing.DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	absPath := filepath.Join(BaseDir, req.Folder)
	if err := os.MkdirAll(absPath, 0755); err != nil {
		http.Error(w, "Failed to create folder", http.StatusInternalServerError)
		return
	}

	go func() {
		for idx, link := range req.ImageLinks {
			utils.DownloadImage(link, absPath, idx)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Images downloading in background"))
}

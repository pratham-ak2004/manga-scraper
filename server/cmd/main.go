package main

import (
	"download-server/cmd/handlers"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/download/", handlers.DownloadHandler)
	http.HandleFunc("/", handlers.FileBrowserHandler)
	http.HandleFunc("/api/downloader/", handlers.DownloadImageHandler)

	fmt.Println("Server starting at:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err.Error())
	}
}

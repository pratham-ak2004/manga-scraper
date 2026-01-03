package main

import (
	"net/http"

	"download-server/cmd/handlers"
	"download-server/cmd/services"
	"download-server/db"
	"download-server/internal/logger"
	"download-server/internal/routes"
	"download-server/internal/server"
)

func bindAllRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", routes.GET(handlers.HomePage))
	mux.HandleFunc("/files", routes.GET(handlers.FilesHandler))
	mux.HandleFunc("/status", routes.GET(handlers.StatusHandler))

	mux.HandleFunc("/download/", routes.GET(handlers.DownloadHandler))
	mux.HandleFunc("/directory/", routes.GET(handlers.GetDirectoryContentHTML))

	mux.HandleFunc("/api/v1/manga", routes.POST(handlers.NewMangaLinkHandler))
	mux.HandleFunc("/api/v1/chapter", routes.POST(handlers.NewChapterLinkHandler))

	assetsFs := http.FileServer(http.Dir("public/assets/"))
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", assetsFs).ServeHTTP(w, r)
	})
}

func main() {
	services.CreateNewCeleryConnection()

	_, err := db.DBConsistentConnection()
	if err != nil {
		logger.Logger.Fatal("Failed to connect to DB: " + err.Error())
	}

	services.Celery.WaitForTaskResultAtStartUp()

	mux := server.GetMux()
	bindAllRoutes(mux)

	s := server.CreateServer("0.0.0.0:8080", mux)

	if err := server.ListenAndServe(s); err != nil {
		logger.Logger.Fatal("Failed to start server: " + err.Error())
	}
}

package main

import (
	"net/http"

	"download-server/cmd/handlers"
	"download-server/cmd/services"
	"download-server/db"
	"download-server/internal/logger"
	"download-server/internal/routes"
	"download-server/internal/server"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func bindAllRoutes(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/dashboard", routes.GET(handlers.DashBoardPage))
	mux.HandleFunc("/dashboard/manga", routes.GET(handlers.MangaListPage))
	mux.HandleFunc("/dashboard/manga/{id}", routes.GET(handlers.MangaPage))
	mux.HandleFunc("/dashboard/explorer/", routes.GET(handlers.FilesHandler))
	mux.HandleFunc("/dashboard/tasks", routes.GET(handlers.TaskPage))
	// TODO: mux.HandleFunc("/manga/read/", routes.GET(handlers.ReadChapter()))

	mux.HandleFunc("/manga/chapters/{id}", routes.GET(handlers.MangaChaptersList))
	mux.HandleFunc("/manga/archives/{id}", routes.GET(handlers.MangaArchivesList))
	mux.HandleFunc("/archives/download/", routes.GET(handlers.DownloadHandler))
	mux.HandleFunc("/directory/", routes.GET(handlers.DirectoryList))
	mux.HandleFunc("/task/status", routes.GET(handlers.TaskDetailsTempl))
	mux.HandleFunc("/task/list", routes.GET(handlers.TaskListTempl))

	mux.HandleFunc("/api/v1/manga", routes.POST(handlers.NewMangaLinkHandler))
	mux.HandleFunc("/api/v1/archive", routes.POST(handlers.ArchiveMangaHandler))
	mux.HandleFunc("/api/v1/eval/manga", routes.POST(handlers.MangaEvalHandler))
	mux.HandleFunc("/api/v1/eval/chapter", routes.POST(handlers.MangaEvalHandler))
	mux.HandleFunc("/api/v1/eval/task", routes.POST(handlers.TaskEvalHandler))

	assetsFs := http.FileServer(http.Dir("public/assets/"))
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", assetsFs).ServeHTTP(w, r)
	})

	mux.HandleFunc("/download/", routes.GET(handlers.DownloadHandler))
	mux.HandleFunc("/", routes.GET(handlers.GlobalHandle))
}

func main() {
	services.CreateNewCeleryConnection()

	_, err := db.DBConsistentConnection()
	if err != nil {
		logger.Logger.Fatal("Failed to connect to DB: " + err.Error())
	}

	go services.Celery.WaitForTaskResultAtStartUp()

	mux := server.GetMux()
	bindAllRoutes(mux)

	s := server.CreateServer("0.0.0.0:8080", mux)

	if err := server.ListenAndServe(s); err != nil {
		logger.Logger.Fatal("Failed to start server: " + err.Error())
	}
}

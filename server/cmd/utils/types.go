package utils

import "download-server/db/generated"

type DownloadRequest struct {
	Folder     string   `json:"folder"`
	ImageLinks []string `json:"image_links"`
}

type DownloadStatus struct {
	Ongoing   int32 `json:"ongoing"`
	Completed int32 `json:"completed"`
	Errored   int32 `json:"errored"`
}

type DashBoardData struct {
	Total struct {
		Manga   int32 `json:"manga"`
		Chapter int64 `json:"chapter"`
		Pages   int64 `json:"pages"`
	} `json:"total"`
	Status struct {
		Completed int32 `json:"completed"`
		Ongoing   int32 `json:"ongoing"`
		Upcoming  int32 `json:"upcomping"`
	}
	Manga []generated.ListMangaDashboardRow `json:"manga"`
}

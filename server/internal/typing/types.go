package typing

type DownloadRequest struct {
	Folder     string   `json:"folder"`
	ImageLinks []string `json:"image_links"`
}

type DownloadStatus struct {
	Ongoing   int32 `json:"ongoing"`
	Completed int32 `json:"completed"`
	Errored   int32 `json:"errored"`
}

type Entry struct {
	Name  string
	Size  string
	Link  string
	IsDir bool
}
package utils

import (
	"strings"
	"time"
)

// FileItem represents a file or directory with extended metadata
type FileItem struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	IsDir       bool      `json:"is_dir"`
	Size        int64     `json:"size"`
	Modified    time.Time `json:"modified"`
	MimeType    string    `json:"mime_type"`
	Icon        string    `json:"icon"`      // Icon type (folder, file, image, archive, etc.)
	Permissions string    `json:"perms"`     // File permissions
	IsHidden    bool      `json:"is_hidden"` // Hidden files
}

// FileExplorerData wraps the file explorer state for rendering
type FileExplorerData struct {
	CurrentPath string       `json:"current_path"`
	Items       FileItemList `json:"items"`
	TotalItems  int          `json:"total_items"`
	TotalSize   int64        `json:"total_size"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
	CanGoUp     bool         `json:"can_go_up"`
}

// FileItemList is a slice of FileItem, used for methods requiring a named type
type FileItemList []FileItem

// Breadcrumb represents a path segment in the breadcrumb navigation
type Breadcrumb struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// DirectoryItem remains for backward compatibility with existing code
type DirectoryItem struct {
	Name  string
	IsDir bool
	URL   string
	Size  int64
}

func (list FileItemList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list FileItemList) Len() int {
	return len(list)
}

func (list FileItemList) Less(i, j int) bool {
	return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
}

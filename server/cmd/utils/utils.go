// Package utils Includes all the utils and typings for handlers
package utils

import (
	"fmt"
	"math/rand/v2"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	BaseDir = "./data/"
	Version = "2.1.0"
)

func GetFolderContent(path string) (FileExplorerData, error) {
	contents, err := os.ReadDir(BaseDir + path)
	if err != nil {
		return FileExplorerData{}, err
	}

	data := FileExplorerData{
		CurrentPath: path,
		Items:       []FileItem{},
		TotalItems:  0,
		TotalSize:   0,
		Breadcrumbs: []Breadcrumb{},
		CanGoUp:     path != "",
	}

	for _, segment := range strings.Split(path, "/") {
		if segment == "" {
			continue
		}
		var crumbPath string
		if len(data.Breadcrumbs) == 0 {
			crumbPath = segment + "/"
		} else {
			crumbPath = data.Breadcrumbs[len(data.Breadcrumbs)-1].Path + segment + "/"
		}
		data.Breadcrumbs = append(data.Breadcrumbs, Breadcrumb{
			Name: segment,
			Path: crumbPath,
		})
	}

	var folders []FileItem
	var files []FileItem

	for _, entry := range contents {
		data.TotalItems += 1

		info, err := entry.Info()
		if err != nil {
			continue
		}
		data.TotalSize += int64(info.Size())

		item := FileItem{
			Name:        info.Name(),
			Path:        "",
			IsDir:       info.IsDir(),
			Size:        info.Size(),
			Modified:    info.ModTime(),
			MimeType:    "application/text",
			Icon:        "file",
			Permissions: info.Mode().String(),
			IsHidden:    strings.HasPrefix(info.Name(), "."),
		}

		if entry.IsDir() {
			item.Path = filepath.Join(path, entry.Name()) + "/"
			item.MimeType = "directory/folder"
			item.Icon = "folder"
			folders = append(folders, item)
		} else {
			item.Path = filepath.Join(path, entry.Name())
			mimeType := mime.TypeByExtension(filepath.Ext(entry.Name()))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			item.MimeType = mimeType

			if strings.HasPrefix(mimeType, "image/") {
				item.Icon = "image"
			} else if strings.HasPrefix(mimeType, "video/") {
				item.Icon = "video"
			} else if strings.HasPrefix(mimeType, "application/zip") || strings.HasPrefix(mimeType, "application/x-rar-compressed") {
				item.Icon = "archive"
			} else {
				item.Icon = "file"
			}

			files = append(files, item)
		}
	}
	sort.Sort(FileItemList(folders))
	sort.Sort(FileItemList(files))
	data.Items = append(folders, files...)

	return data, nil
}

func FormatSize(size int64, isDir bool) string {
	if isDir {
		return "-"
	}
	switch {
	case size >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	case size >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func WithTicker(action func() bool) {
	waitTime := 5
	ticker := time.NewTicker(time.Duration(waitTime) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		completed := action()

		if completed {
			break
		} else {
			waitTime += rand.IntN(5)
			ticker.Reset(time.Duration(waitTime) * time.Second)
		}
	}
}

func CreateToast(w http.ResponseWriter, toastType string, body string) {
	w.Header().Set("HXToaster-Body", body)
	w.Header().Set("HXToaster-Type", toastType)
}

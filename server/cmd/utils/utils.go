// Package utils Includes all the utils and typings for handlers
package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

const BaseDir = "./data"

var (
	status DownloadStatus
	mu     sync.Mutex
)

func GetFolderContent(path string) ([]DirectoryItem, error) {
	contents, err := os.ReadDir(BaseDir + path)
	if err != nil {
		return []DirectoryItem{}, err
	}

	var directoryItems []DirectoryItem
	for _, entry := range contents {
		info, err := entry.Info()
		var size int64

		if err != nil {
			size = 0
		} else {
			size = info.Size()
		}

		var url string
		if entry.IsDir() {
			url = path + entry.Name() + "/"
		} else {
			url = path + entry.Name()
		}

		newItem := DirectoryItem{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			URL:   url,
			Size:  size,
		}

		directoryItems = append(directoryItems, newItem)
	}

	var folders []DirectoryItem
	var files []DirectoryItem
	for _, item := range directoryItems {
		if item.IsDir {
			folders = append(folders, item)
		} else {
			files = append(files, item)
		}
	}

	return append(folders, files...), nil
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

func DownloadImage(url, folder string, idx int) {
	atomic.AddInt32(&status.Ongoing, 1)
	defer atomic.AddInt32(&status.Ongoing, -1)

	mu.Lock()
	resp, err := http.Get(url)
	mu.Unlock()

	if err != nil {
		fmt.Printf("Failed to download %s: %v\n", url, err)
		atomic.AddInt32(&status.Errored, 1)
		return
	}
	defer resp.Body.Close()

	os.MkdirAll(folder, os.ModePerm)
	filename := filepath.Join(folder, fmt.Sprintf("%03d.jpg", idx))
	out, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create file %s: %v\n", filename, err)
		atomic.AddInt32(&status.Errored, 1)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save %s: %v\n", filename, err)
		atomic.AddInt32(&status.Errored, 1)
		return
	}
	atomic.AddInt32(&status.Completed, 1)
}

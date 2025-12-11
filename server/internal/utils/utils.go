package utils

import (
	"download-server/internal/typing"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var (
	status typing.DownloadStatus
	mu     sync.Mutex
)

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

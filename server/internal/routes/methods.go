package routes

import (
	"net/http"

	logger "download-server/internal/logger"
)

var colors = logger.Colors

func FileServerMiddleware(next http.Handler) http.Handler {
	return logger.Timer(colors["blue"]+"FileServer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func GET(next http.HandlerFunc) http.HandlerFunc {
	return logger.Timer(colors["magenta"]+"GET", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		} else {
			next(w, r)
		}
	})
}

func POST(next http.HandlerFunc) http.HandlerFunc {
	return logger.Timer(colors["cyan"]+"POST", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		} else {
			next(w, r)
		}
	})
}

func DELETE(next http.HandlerFunc) http.HandlerFunc {
	return logger.Timer(colors["cyan"]+"DELETE", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		} else {
			next(w, r)
		}
	})
}

func ServFile(path string) http.HandlerFunc {
	return logger.Timer(colors["green"]+"File", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		} else {
			http.ServeFile(w, r, path)
		}
	})
}

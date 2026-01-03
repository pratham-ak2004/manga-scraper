package logger

import (
	"net/http"
	"time"
)

var logger = Logger
var colors = Colors

func Timer(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// cors.CorsMiddleware(next).ServeHTTP(w, r)
		next(w, r)

		end := time.Since(start).Milliseconds()
		var timeColor string
		if end < 100 {
			timeColor = colors["green"]
		} else if end < 500 {
			timeColor = colors["yellow"]
		} else {
			timeColor = colors["red"]
		}

		logger.Printf("%s %d%s %s %s", timeColor, end, "ms", method, colors["reset"]+r.URL.Path)
	}
}

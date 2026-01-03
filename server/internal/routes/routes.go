package routes

import (
	"net/http"
)

func BindRoutes(mux *http.ServeMux, bindFunc func(mux *http.ServeMux)) {
	bindFunc(mux)
}

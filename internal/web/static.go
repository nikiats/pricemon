package web

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

var staticFiles = os.DirFS("internal/web/static")

func StaticHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", pageHandler("pages/index.html"))
	mux.HandleFunc("GET /login", pageHandler("pages/login.html"))
	mux.HandleFunc("GET /register", pageHandler("pages/register.html"))
	mux.HandleFunc("GET /css/", assetHandler("/css/", "css"))
	mux.HandleFunc("GET /js/", assetHandler("/js/", "js"))

	return mux
}

func pageHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFiles, name)
	}
}

func assetHandler(prefix, directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, prefix)
		if !fs.ValidPath(name) {
			http.NotFound(w, r)
			return
		}

		http.ServeFileFS(w, r, staticFiles, path.Join(directory, name))
	}
}

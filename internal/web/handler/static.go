package handler

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed static
var staticFiles embed.FS

func pageHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFiles, path.Join("static", name))
	}
}

func assetHandler(prefix, directory string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, prefix)
		if !fs.ValidPath(name) {
			http.NotFound(w, r)
			return
		}

		http.ServeFileFS(w, r, staticFiles, path.Join("static", directory, name))
	}
}

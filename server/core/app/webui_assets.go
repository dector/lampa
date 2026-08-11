package app

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const WebUIAssetsPath = "/assets/"

//go:embed assets/*
var webUIAssets embed.FS

// NewWebUIAssetHandler serves embedded Web UI assets under /assets/.
func NewWebUIAssetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cleanPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		if !strings.HasPrefix(cleanPath, WebUIAssetsPath) {
			http.NotFound(w, r)
			return
		}

		assetPath := strings.TrimPrefix(cleanPath, "/")
		info, err := fs.Stat(webUIAssets, assetPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeFileFS(w, r, webUIAssets, assetPath)
	}
}

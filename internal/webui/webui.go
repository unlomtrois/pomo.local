// Package webui embeds the compiled Svelte dashboard and serves it as a
// single-page app. The assets are synced into ./dist by `make ui` (from the
// SvelteKit static build) and compiled into the binary via go:embed, so the
// daemon serves the UI same-origin with the API — no separate web server.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// all: so dot/underscore files (e.g. SvelteKit's _app/) are included.
//
//go:embed all:dist
var distFS embed.FS

// Handler serves the embedded UI with SPA fallback: unknown client-side routes
// (e.g. /calendar) return index.html so the app can route them. If only the
// placeholder is embedded (UI not built), it serves a short notice.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return notBuilt("embed error: " + err.Error())
	}
	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return notBuilt("UI not built — run `make ui` then rebuild")
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" || p == "." {
			serveIndex(w, index)
			return
		}
		if f, err := sub.Open(p); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		serveIndex(w, index) // SPA fallback
	})
}

func serveIndex(w http.ResponseWriter, index []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(index)
}

func notBuilt(msg string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("pomo daemon: " + msg + "\n"))
	})
}

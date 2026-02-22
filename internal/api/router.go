package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/api/handlers"
	apimw "github.com/lamngockhuong/dbsight/internal/api/middleware"
)

func NewRouter(app *App, staticFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(apimw.Logger)
	r.Use(apimw.Recovery)
	r.Use(corsMiddleware)

	h := handlers.New(app.Store, app.CryptoKey, app.NewAdapter)

	r.Route("/api", func(r chi.Router) {
		r.Route("/connections", func(r chi.Router) {
			r.Get("/", h.ListConnections)
			r.Post("/", h.CreateConnection)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetConnection)
				r.Put("/", h.UpdateConnection)
				r.Delete("/", h.DeleteConnection)
				r.Post("/test", h.TestConnection)
				r.Get("/queries", h.ListQueries)
				r.Get("/queries/stream", h.StreamQueries)
				r.Get("/queries/history", h.ListQueryHistory)
			})
		})
		r.Post("/paste/queries", h.ParseSlowLog)
	})

	r.Handle("/*", spaHandler(staticFS))
	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		f, err := staticFS.Open(path)
		if err != nil {
			r.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}

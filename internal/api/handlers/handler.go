package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/store"
)

type Handler struct {
	store      store.Store
	cryptoKey  []byte
	newAdapter func(string) (adapter.DBAnalyzer, error)
}

func New(s store.Store, key []byte, f func(string) (adapter.DBAnalyzer, error)) *Handler {
	return &Handler{store: s, cryptoKey: key, newAdapter: f}
}

func jsonOK(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/crypto"
)

func (h *Handler) RunExplain(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Query       string `json:"query"`
		AnalyzeMode bool   `json:"analyze_mode"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		jsonError(w, "query is required", http.StatusBadRequest)
		return
	}

	conn, err := h.store.GetConnection(r.Context(), id)
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}

	dsn, err := crypto.Decrypt(h.cryptoKey, conn.EncryptedDSN)
	if err != nil {
		jsonError(w, "decrypt error", http.StatusInternalServerError)
		return
	}

	a, err := h.newAdapter(conn.DBType)
	if err != nil {
		jsonError(w, "unsupported db type", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := a.Connect(ctx, string(dsn)); err != nil {
		jsonError(w, "connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer a.Close()

	plan, err := a.GetExplainPlan(ctx, req.Query, adapter.QueryOpts{AnalyzeMode: req.AnalyzeMode})
	if err != nil {
		jsonError(w, "explain failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, plan, http.StatusOK)
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/crypto"
	"github.com/lamngockhuong/dbsight/internal/models"
)

func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	conns, err := h.store.ListConnections(r.Context())
	if err != nil {
		jsonError(w, "failed to list connections", http.StatusInternalServerError)
		return
	}
	if conns == nil {
		conns = []models.Connection{}
	}
	jsonOK(w, conns, http.StatusOK)
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		DBType string `json:"db_type"`
		DSN    string `json:"dsn"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.DSN == "" {
		jsonError(w, "name and dsn are required", http.StatusBadRequest)
		return
	}
	if req.DBType == "" {
		req.DBType = "postgres"
	}

	encrypted, err := crypto.Encrypt(h.cryptoKey, []byte(req.DSN))
	if err != nil {
		jsonError(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	conn := &models.Connection{
		Name:         req.Name,
		DBType:       req.DBType,
		EncryptedDSN: encrypted,
	}
	if err := h.store.CreateConnection(r.Context(), conn); err != nil {
		jsonError(w, "failed to create connection", http.StatusInternalServerError)
		return
	}
	jsonOK(w, conn, http.StatusCreated)
}

func (h *Handler) GetConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	conn, err := h.store.GetConnection(r.Context(), id)
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}
	jsonOK(w, conn, http.StatusOK)
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct {
		Name   string `json:"name"`
		DBType string `json:"db_type"`
		DSN    string `json:"dsn"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	conn, err := h.store.GetConnection(r.Context(), id)
	if err != nil {
		jsonError(w, "connection not found", http.StatusNotFound)
		return
	}
	if req.Name != "" {
		conn.Name = req.Name
	}
	if req.DBType != "" {
		conn.DBType = req.DBType
	}
	if req.DSN != "" {
		encrypted, err := crypto.Encrypt(h.cryptoKey, []byte(req.DSN))
		if err != nil {
			jsonError(w, "encryption failed", http.StatusInternalServerError)
			return
		}
		conn.EncryptedDSN = encrypted
	}
	if err := h.store.UpdateConnection(r.Context(), conn); err != nil {
		jsonError(w, "failed to update connection", http.StatusInternalServerError)
		return
	}
	jsonOK(w, conn, http.StatusOK)
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteConnection(r.Context(), id); err != nil {
		jsonError(w, "failed to delete connection", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
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
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := a.Connect(ctx, string(dsn)); err != nil {
		jsonError(w, "connection failed", http.StatusBadGateway)
		return
	}
	defer a.Close()
	jsonOK(w, map[string]any{"latency_ms": time.Since(start).Milliseconds()}, http.StatusOK)
}

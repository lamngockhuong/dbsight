package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/models"
)

func (h *Handler) ListQueries(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	snaps, err := h.store.ListQuerySnapshots(r.Context(), id, 2)
	if err != nil {
		jsonError(w, "failed to list queries", http.StatusInternalServerError)
		return
	}
	if len(snaps) == 0 {
		jsonOK(w, []any{}, http.StatusOK)
		return
	}

	latest := snaps[0]
	var deltas []models.QueryDelta
	if len(snaps) == 2 {
		deltas = computeDeltas(latest, snaps[1])
	} else {
		for _, q := range latest.Queries {
			deltas = append(deltas, models.QueryDelta{SlowQuery: q})
		}
	}
	if deltas == nil {
		deltas = []models.QueryDelta{}
	}
	jsonOK(w, deltas, http.StatusOK)
}

func (h *Handler) StreamQueries(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		jsonError(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			snaps, err := h.store.ListQuerySnapshots(r.Context(), id, 2)
			if err != nil || len(snaps) == 0 {
				continue
			}
			var deltas []models.QueryDelta
			if len(snaps) >= 2 {
				deltas = computeDeltas(snaps[0], snaps[1])
			} else {
				for _, q := range snaps[0].Queries {
					deltas = append(deltas, models.QueryDelta{SlowQuery: q})
				}
			}
			data, _ := json.Marshal(deltas)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *Handler) ListQueryHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	snaps, err := h.store.ListQuerySnapshots(r.Context(), id, limit)
	if err != nil {
		jsonError(w, "failed to list history", http.StatusInternalServerError)
		return
	}
	if snaps == nil {
		snaps = []models.QuerySnapshot{}
	}
	jsonOK(w, snaps, http.StatusOK)
}

func computeDeltas(curr, prev models.QuerySnapshot) []models.QueryDelta {
	prevMap := make(map[string]models.SlowQuery, len(prev.Queries))
	for _, q := range prev.Queries {
		prevMap[q.QueryID] = q
	}
	period := curr.CapturedAt.Sub(prev.CapturedAt).Seconds()
	var out []models.QueryDelta
	for _, q := range curr.Queries {
		d := models.QueryDelta{SlowQuery: q, PeriodSecs: period}
		if p, ok := prevMap[q.QueryID]; ok {
			d.CallsDelta = q.Calls - p.Calls
			d.TotalExecDelta = q.TotalExecMs - p.TotalExecMs
			d.MeanExecDelta = q.MeanExecMs - p.MeanExecMs
		}
		out = append(out, d)
	}
	return out
}

package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lamngockhuong/dbsight/internal/models"
)

var slowLogPattern = regexp.MustCompile(`duration:\s*([\d.]+)\s*ms\s+statement:\s*(.+)`)

func (h *Handler) ParseSlowLog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LogText string `json:"log_text"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 10<<20)).Decode(&req); err != nil {
		jsonError(w, "invalid request (max 10MB)", http.StatusBadRequest)
		return
	}
	queries := parsePostgresSlowLog(req.LogText)
	jsonOK(w, queries, http.StatusOK)
}

func parsePostgresSlowLog(text string) []models.SlowQuery {
	type agg struct {
		query   string
		totalMs float64
		calls   int64
	}
	grouped := map[string]*agg{}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		matches := slowLogPattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		durationMs, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			continue
		}
		query := strings.TrimSpace(matches[2])
		fingerprint := normalizeQuery(query)

		if g, ok := grouped[fingerprint]; ok {
			g.totalMs += durationMs
			g.calls++
		} else {
			grouped[fingerprint] = &agg{query: query, totalMs: durationMs, calls: 1}
		}
	}

	now := time.Now()
	var result []models.SlowQuery
	for fp, g := range grouped {
		result = append(result, models.SlowQuery{
			QueryID:     fp,
			Query:       g.query,
			Calls:       g.calls,
			TotalExecMs: g.totalMs,
			MeanExecMs:  g.totalMs / float64(g.calls),
			SnapshotAt:  now,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalExecMs > result[j].TotalExecMs
	})
	return result
}

var (
	numPattern = regexp.MustCompile(`\b\d+\b`)
	strPattern = regexp.MustCompile(`'[^']*'`)
)

func normalizeQuery(q string) string {
	q = strings.ToLower(q)
	q = strings.Join(strings.Fields(q), " ")
	q = numPattern.ReplaceAllString(q, "?")
	q = strPattern.ReplaceAllString(q, "?")
	return q
}

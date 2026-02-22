package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/crypto"
	"github.com/lamngockhuong/dbsight/internal/models"
)

func (h *Handler) GetIndexAnalysis(w http.ResponseWriter, r *http.Request) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := a.Connect(ctx, string(dsn)); err != nil {
		jsonError(w, "connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer a.Close()

	allStats, _ := a.GetIndexStats(ctx)
	tableStats, _ := a.GetTableStats(ctx)

	// Need type assertion for duplicate index support
	var duplicates []models.DuplicateIndex
	if pa, ok := a.(interface {
		GetDuplicateIndexes(context.Context) ([]models.DuplicateIndex, error)
	}); ok {
		duplicates, _ = pa.GetDuplicateIndexes(ctx)
	}

	var unused []models.IndexStat
	for _, s := range allStats {
		if s.IsUnused {
			unused = append(unused, s)
		}
	}

	var highSeqScan []models.TableStat
	for _, t := range tableStats {
		if t.SeqScans > 100 && t.NLiveTup > 1000 {
			highSeqScan = append(highSeqScan, t)
		}
	}

	result := &models.IndexAnalysisResult{
		UnusedIndexes:     unused,
		MissingCandidates: highSeqScan,
		DuplicateIndexes:  duplicates,
		CapturedAt:        time.Now(),
	}

	// Ensure non-nil slices for JSON
	if result.UnusedIndexes == nil {
		result.UnusedIndexes = []models.IndexStat{}
	}
	if result.MissingCandidates == nil {
		result.MissingCandidates = []models.TableStat{}
	}
	if result.DuplicateIndexes == nil {
		result.DuplicateIndexes = []models.DuplicateIndex{}
	}

	result.Recommendations = computeRecommendations(result)
	jsonOK(w, result, http.StatusOK)
}

func computeRecommendations(result *models.IndexAnalysisResult) []models.Recommendation {
	var recs []models.Recommendation

	for _, idx := range result.UnusedIndexes {
		if idx.IndexSizeB > 1024*1024 { // > 1MB worth dropping
			recs = append(recs, models.Recommendation{
				Type:       "drop_unused",
				SchemaName: idx.SchemaName,
				TableName:  idx.TableName,
				IndexName:  idx.IndexName,
				Description: fmt.Sprintf("Index %s has 0 scans and uses %s",
					idx.IndexName, formatBytes(idx.IndexSizeB)),
				SQL:      fmt.Sprintf("DROP INDEX CONCURRENTLY %s.%s;", idx.SchemaName, idx.IndexName),
				Severity: "medium",
			})
		}
	}

	for _, t := range result.MissingCandidates {
		if t.SeqScans > 1000 && t.NLiveTup > 10000 {
			recs = append(recs, models.Recommendation{
				Type:       "add_index",
				SchemaName: t.SchemaName,
				TableName:  t.TableName,
				Description: fmt.Sprintf("Table %s has %d seq scans on %d rows — consider adding indexes",
					t.TableName, t.SeqScans, t.NLiveTup),
				Severity: "high",
			})
		}
	}

	for _, dup := range result.DuplicateIndexes {
		recs = append(recs, models.Recommendation{
			Type:        "drop_duplicate",
			TableName:   dup.TableName,
			IndexName:   dup.Index2,
			Description: fmt.Sprintf("%s is a duplicate of %s", dup.Index2, dup.Index1),
			SQL:         fmt.Sprintf("DROP INDEX CONCURRENTLY %s;", dup.Index2),
			Severity:    "medium",
		})
	}

	if recs == nil {
		recs = []models.Recommendation{}
	}
	return recs
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

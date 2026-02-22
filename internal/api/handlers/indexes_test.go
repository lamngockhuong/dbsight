package handlers

import (
	"testing"

	"github.com/lamngockhuong/dbsight/internal/models"
)

func TestComputeRecommendations_UnusedLargeIndex(t *testing.T) {
	result := &models.IndexAnalysisResult{
		UnusedIndexes: []models.IndexStat{
			{SchemaName: "public", TableName: "users", IndexName: "idx_old", IndexSizeB: 2 << 20, IsUnused: true},
		},
		MissingCandidates: []models.TableStat{},
		DuplicateIndexes:  []models.DuplicateIndex{},
	}

	recs := computeRecommendations(result)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].Type != "drop_unused" {
		t.Errorf("expected type drop_unused, got %s", recs[0].Type)
	}
	if recs[0].Severity != "medium" {
		t.Errorf("expected severity medium, got %s", recs[0].Severity)
	}
}

func TestComputeRecommendations_SmallUnusedIndex_NoRec(t *testing.T) {
	result := &models.IndexAnalysisResult{
		UnusedIndexes: []models.IndexStat{
			{SchemaName: "public", TableName: "t", IndexName: "idx_small", IndexSizeB: 512, IsUnused: true},
		},
		MissingCandidates: []models.TableStat{},
		DuplicateIndexes:  []models.DuplicateIndex{},
	}

	recs := computeRecommendations(result)
	if len(recs) != 0 {
		t.Fatalf("expected 0 recommendations for small index, got %d", len(recs))
	}
}

func TestComputeRecommendations_HighSeqScan(t *testing.T) {
	result := &models.IndexAnalysisResult{
		UnusedIndexes: []models.IndexStat{},
		MissingCandidates: []models.TableStat{
			{SchemaName: "public", TableName: "orders", SeqScans: 5000, NLiveTup: 50000},
		},
		DuplicateIndexes: []models.DuplicateIndex{},
	}

	recs := computeRecommendations(result)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].Type != "add_index" {
		t.Errorf("expected type add_index, got %s", recs[0].Type)
	}
	if recs[0].Severity != "high" {
		t.Errorf("expected severity high, got %s", recs[0].Severity)
	}
}

func TestComputeRecommendations_Duplicates(t *testing.T) {
	result := &models.IndexAnalysisResult{
		UnusedIndexes:     []models.IndexStat{},
		MissingCandidates: []models.TableStat{},
		DuplicateIndexes: []models.DuplicateIndex{
			{TableName: "users", Index1: "idx_a", Index2: "idx_b", IndexDef: "CREATE INDEX ..."},
		},
	}

	recs := computeRecommendations(result)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].Type != "drop_duplicate" {
		t.Errorf("expected type drop_duplicate, got %s", recs[0].Type)
	}
}

func TestComputeRecommendations_Empty(t *testing.T) {
	result := &models.IndexAnalysisResult{
		UnusedIndexes:     []models.IndexStat{},
		MissingCandidates: []models.TableStat{},
		DuplicateIndexes:  []models.DuplicateIndex{},
	}

	recs := computeRecommendations(result)
	if len(recs) != 0 {
		t.Fatalf("expected 0 recommendations, got %d", len(recs))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{2048, "2.0 KB"},
		{1572864, "1.5 MB"},
		{2147483648, "2.0 GB"},
	}
	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

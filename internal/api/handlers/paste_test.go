package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamngockhuong/dbsight/internal/models"
)

func TestParseSlowLog(t *testing.T) {
	h := newTestHandler()
	logText := `2026-02-22 LOG: duration: 150.5 ms  statement: SELECT * FROM users WHERE id = 1
2026-02-22 LOG: duration: 200.3 ms  statement: SELECT * FROM users WHERE id = 2
2026-02-22 LOG: duration: 50.0 ms  statement: INSERT INTO logs VALUES (1, 'test')`

	body, _ := json.Marshal(map[string]string{"log_text": logText})
	req := httptest.NewRequest(http.MethodPost, "/api/paste/queries", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.ParseSlowLog(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var queries []models.SlowQuery
	if err := json.NewDecoder(w.Body).Decode(&queries); err != nil {
		t.Fatal(err)
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 grouped queries, got %d", len(queries))
	}
	// First result should be highest total_exec_ms (SELECT grouped: 150.5+200.3=350.8)
	if queries[0].TotalExecMs < queries[1].TotalExecMs {
		t.Error("expected results sorted by total_exec_ms descending")
	}
}

func TestParseSlowLog_EmptyInput(t *testing.T) {
	h := newTestHandler()
	body, _ := json.Marshal(map[string]string{"log_text": ""})
	req := httptest.NewRequest(http.MethodPost, "/api/paste/queries", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.ParseSlowLog(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SELECT * FROM users WHERE id = 42", "select * from users where id = ?"},
		{"INSERT INTO logs VALUES ('hello', 123)", "insert into logs values (?, ?)"},
		{"SELECT  *  FROM  users", "select * from users"},
	}
	for _, tt := range tests {
		result := normalizeQuery(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeQuery(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/models"
)

// mockStore implements store.Store for testing handlers.
type mockStore struct {
	connections map[int64]*models.Connection
	nextID      int64
}

func newMockStore() *mockStore {
	return &mockStore{connections: make(map[int64]*models.Connection), nextID: 1}
}

func (m *mockStore) Ping(_ context.Context) error { return nil }

func (m *mockStore) CreateConnection(_ context.Context, c *models.Connection) error {
	c.ID = m.nextID
	m.nextID++
	m.connections[c.ID] = c
	return nil
}

func (m *mockStore) GetConnection(_ context.Context, id int64) (*models.Connection, error) {
	c, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockStore) ListConnections(_ context.Context) ([]models.Connection, error) {
	var result []models.Connection
	for _, c := range m.connections {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockStore) UpdateConnection(_ context.Context, c *models.Connection) error {
	if _, ok := m.connections[c.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.connections[c.ID] = c
	return nil
}

func (m *mockStore) DeleteConnection(_ context.Context, id int64) error {
	delete(m.connections, id)
	return nil
}

func (m *mockStore) SaveQuerySnapshot(_ context.Context, _ *models.QuerySnapshot) error {
	return nil
}

func (m *mockStore) GetLatestQuerySnapshot(_ context.Context, _ int64) (*models.QuerySnapshot, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockStore) ListQuerySnapshots(_ context.Context, _ int64, _ int) ([]models.QuerySnapshot, error) {
	return nil, nil
}

func (m *mockStore) SaveIndexStatsSnapshot(_ context.Context, _ int64, _ []models.IndexStat) error {
	return nil
}

func (m *mockStore) GetLatestIndexStats(_ context.Context, _ int64) ([]models.IndexStat, error) {
	return nil, nil
}

// testCryptoKey is a valid 32-byte key for testing.
var testCryptoKey = []byte("01234567890123456789012345678901")

func failAdapter(_ string) (adapter.DBAnalyzer, error) {
	return nil, fmt.Errorf("no real adapter in tests")
}

func newTestHandler() *Handler {
	return New(newMockStore(), testCryptoKey, failAdapter)
}

// withChiURLParam sets a chi URL param on the request context.
func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestListConnections_Empty(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/connections", nil)
	w := httptest.NewRecorder()

	h.ListConnections(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var conns []models.Connection
	if err := json.NewDecoder(w.Body).Decode(&conns); err != nil {
		t.Fatal(err)
	}
	if len(conns) != 0 {
		t.Fatalf("expected empty list, got %d items", len(conns))
	}
}

func TestCreateConnection(t *testing.T) {
	h := newTestHandler()
	body := `{"name":"test-db","db_type":"postgres","dsn":"postgres://localhost/test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/connections", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateConnection(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var conn models.Connection
	if err := json.NewDecoder(w.Body).Decode(&conn); err != nil {
		t.Fatal(err)
	}
	if conn.Name != "test-db" {
		t.Errorf("expected name test-db, got %s", conn.Name)
	}
	if conn.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateConnection_MissingFields(t *testing.T) {
	h := newTestHandler()
	body := `{"name":"","dsn":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/connections", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateConnection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetConnection_NotFound(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/connections/999", nil)
	req = withChiURLParam(req, "id", "999")
	w := httptest.NewRecorder()

	h.GetConnection(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetConnection_InvalidID(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/connections/abc", nil)
	req = withChiURLParam(req, "id", "abc")
	w := httptest.NewRecorder()

	h.GetConnection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteConnection(t *testing.T) {
	h := newTestHandler()

	// Create first
	body := `{"name":"del-me","dsn":"postgres://localhost/test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/connections", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.CreateConnection(w, req)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/connections/1", nil)
	req = withChiURLParam(req, "id", "1")
	w = httptest.NewRecorder()
	h.DeleteConnection(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

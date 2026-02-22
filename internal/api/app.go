package api

import (
	"github.com/lamngockhuong/dbsight/internal/adapter"
	"github.com/lamngockhuong/dbsight/internal/store"
)

type App struct {
	Store      store.Store
	CryptoKey  []byte
	NewAdapter func(dbType string) (adapter.DBAnalyzer, error)
}

package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear env vars to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("ENCRYPTION_KEY")
	os.Unsetenv("WORKER_INTERVAL_SECS")

	cfg := Load()
	if cfg.Port != "42198" {
		t.Errorf("expected port 42198, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("expected empty DATABASE_URL, got %s", cfg.DatabaseURL)
	}
	if cfg.WorkerInterval != 30 {
		t.Errorf("expected worker interval 30, got %d", cfg.WorkerInterval)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("ENCRYPTION_KEY", "abc123")
	t.Setenv("WORKER_INTERVAL_SECS", "60")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected DATABASE_URL postgres://localhost/test, got %s", cfg.DatabaseURL)
	}
	if cfg.EncryptionKey != "abc123" {
		t.Errorf("expected ENCRYPTION_KEY abc123, got %s", cfg.EncryptionKey)
	}
	if cfg.WorkerInterval != 60 {
		t.Errorf("expected worker interval 60, got %d", cfg.WorkerInterval)
	}
}

func TestLoad_InvalidWorkerInterval(t *testing.T) {
	t.Setenv("WORKER_INTERVAL_SECS", "not-a-number")

	cfg := Load()
	if cfg.WorkerInterval != 30 {
		t.Errorf("expected fallback 30 for invalid interval, got %d", cfg.WorkerInterval)
	}
}

package adapter

import (
	"testing"
)

func TestNewAdapter_Postgres(t *testing.T) {
	a, err := NewAdapter("postgres")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil adapter")
	}
}

func TestNewAdapter_Unsupported(t *testing.T) {
	_, err := NewAdapter("mysql")
	if err == nil {
		t.Fatal("expected error for unsupported db type")
	}
}

func TestNewAdapter_Empty(t *testing.T) {
	_, err := NewAdapter("")
	if err == nil {
		t.Fatal("expected error for empty db type")
	}
}

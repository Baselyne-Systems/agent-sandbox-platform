package workspace

import (
	"context"
	"errors"
	"testing"
)

func TestLocalSnapshotStore_SaveReturnsNotImplemented(t *testing.T) {
	store, err := NewLocalSnapshotStore(t.TempDir())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	_, err = store.SaveSnapshot(context.Background(), "ws-123")
	if !errors.Is(err, ErrSnapshotNotImplemented) {
		t.Errorf("expected ErrSnapshotNotImplemented, got: %v", err)
	}
}

func TestLocalSnapshotStore_LoadReturnsNotImplemented(t *testing.T) {
	store, err := NewLocalSnapshotStore(t.TempDir())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	err = store.LoadSnapshot(context.Background(), "snap-123")
	if !errors.Is(err, ErrSnapshotNotImplemented) {
		t.Errorf("expected ErrSnapshotNotImplemented, got: %v", err)
	}
}

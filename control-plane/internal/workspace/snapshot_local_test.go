package workspace

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSnapshotStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalSnapshotStore(dir)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	ctx := context.Background()
	snapshotID, err := store.SaveSnapshot(ctx, "ws-123")
	if err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	if snapshotID == "" {
		t.Fatal("expected non-empty snapshot ID")
	}

	// Verify metadata file.
	metaPath := filepath.Join(dir, snapshotID, "metadata.json")
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	var meta snapshotMetadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if meta.WorkspaceID != "ws-123" {
		t.Errorf("expected workspace_id 'ws-123', got %q", meta.WorkspaceID)
	}
	if meta.SnapshotID != snapshotID {
		t.Errorf("expected snapshot_id %q, got %q", snapshotID, meta.SnapshotID)
	}

	// Verify tarball placeholder exists.
	tarPath := filepath.Join(dir, snapshotID, "fs.tar.gz")
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		t.Error("expected fs.tar.gz to exist")
	}

	// Load should succeed.
	if err := store.LoadSnapshot(ctx, snapshotID); err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
}

func TestLocalSnapshotStore_LoadNotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalSnapshotStore(dir)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	err = store.LoadSnapshot(context.Background(), "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestLocalSnapshotStore_MultipleSnapshots(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalSnapshotStore(dir)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	ctx := context.Background()

	id1, err := store.SaveSnapshot(ctx, "ws-1")
	if err != nil {
		t.Fatalf("save 1: %v", err)
	}
	id2, err := store.SaveSnapshot(ctx, "ws-2")
	if err != nil {
		t.Fatalf("save 2: %v", err)
	}
	if id1 == id2 {
		t.Error("expected unique snapshot IDs")
	}

	// Both should load.
	if err := store.LoadSnapshot(ctx, id1); err != nil {
		t.Errorf("load 1: %v", err)
	}
	if err := store.LoadSnapshot(ctx, id2); err != nil {
		t.Errorf("load 2: %v", err)
	}
}

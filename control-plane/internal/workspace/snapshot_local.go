package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// snapshotMetadata is persisted alongside each snapshot.
type snapshotMetadata struct {
	SnapshotID  string    `json:"snapshot_id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// LocalSnapshotStore stores workspace snapshots on the local filesystem.
// Each snapshot is a directory under baseDir containing a metadata JSON file.
// In production, the snapshot directory would also contain a filesystem tarball
// exported from the sandbox container.
type LocalSnapshotStore struct {
	baseDir string
}

// NewLocalSnapshotStore creates a snapshot store rooted at baseDir.
// The directory is created if it does not exist.
func NewLocalSnapshotStore(baseDir string) (*LocalSnapshotStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot base dir: %w", err)
	}
	return &LocalSnapshotStore{baseDir: baseDir}, nil
}

func (s *LocalSnapshotStore) SaveSnapshot(_ context.Context, workspaceID string) (string, error) {
	snapshotID := uuid.New().String()
	dir := filepath.Join(s.baseDir, snapshotID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create snapshot dir: %w", err)
	}

	meta := snapshotMetadata{
		SnapshotID:  snapshotID,
		WorkspaceID: workspaceID,
		CreatedAt:   time.Now(),
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal metadata: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), metaBytes, 0o644); err != nil {
		return "", fmt.Errorf("write metadata: %w", err)
	}

	// In production: export container filesystem here (e.g., docker export → fs.tar.gz).
	// For now, create a placeholder marker file.
	if err := os.WriteFile(filepath.Join(dir, "fs.tar.gz"), []byte{}, 0o644); err != nil {
		return "", fmt.Errorf("write placeholder tarball: %w", err)
	}

	return snapshotID, nil
}

func (s *LocalSnapshotStore) LoadSnapshot(_ context.Context, snapshotID string) error {
	dir := filepath.Join(s.baseDir, snapshotID)
	metaPath := filepath.Join(dir, "metadata.json")

	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	// In production: import fs.tar.gz into a new container here.
	// Verify the tarball exists.
	tarPath := filepath.Join(dir, "fs.tar.gz")
	if _, err := os.Stat(tarPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot tarball not found for %s", snapshotID)
	}

	return nil
}

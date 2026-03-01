package workspace

import (
	"context"
	"errors"
	"fmt"
)

// ErrSnapshotNotImplemented is returned by snapshot operations that require
// Docker container filesystem export/import, which is not yet implemented.
// A real implementation would add ExportSandbox/ImportSandbox RPCs to the
// Host Agent proto and use bollard's export_container/import_image APIs.
var ErrSnapshotNotImplemented = errors.New("snapshot: container filesystem export not yet implemented")

// LocalSnapshotStore stores workspace snapshots on the local filesystem.
type LocalSnapshotStore struct {
	baseDir string
}

// NewLocalSnapshotStore creates a snapshot store rooted at baseDir.
func NewLocalSnapshotStore(baseDir string) (*LocalSnapshotStore, error) {
	return &LocalSnapshotStore{baseDir: baseDir}, nil
}

func (s *LocalSnapshotStore) SaveSnapshot(_ context.Context, workspaceID string) (string, error) {
	return "", fmt.Errorf("save snapshot for workspace %s: %w", workspaceID, ErrSnapshotNotImplemented)
}

func (s *LocalSnapshotStore) LoadSnapshot(_ context.Context, snapshotID string) error {
	return fmt.Errorf("load snapshot %s: %w", snapshotID, ErrSnapshotNotImplemented)
}

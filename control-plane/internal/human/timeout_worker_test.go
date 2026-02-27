package human

import (
	"context"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

func TestTimeoutWorker_ExpiresRequests(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	ctx := context.Background()
	req, err := svc.CreateRequest(ctx, "ws-1", "agent-1", "question?", nil, "", 3600, "", "", "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	// Manually set expires_at to the past.
	pastTime := time.Now().Add(-1 * time.Minute)
	repo.requests[req.ID].ExpiresAt = &pastTime

	worker := NewTimeoutWorker(repo, TimeoutWorkerConfig{Interval: time.Second})
	worker.sweep(ctx)

	got := repo.requests[req.ID]
	if got.Status != models.HumanRequestStatusExpired {
		t.Errorf("expected status expired, got %q", got.Status)
	}
}

func TestTimeoutWorker_SkipsFutureRequests(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	ctx := context.Background()
	req, err := svc.CreateRequest(ctx, "ws-1", "agent-1", "question?", nil, "", 3600, "", "", "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	worker := NewTimeoutWorker(repo, TimeoutWorkerConfig{Interval: time.Second})
	worker.sweep(ctx)

	got := repo.requests[req.ID]
	if got.Status != models.HumanRequestStatusPending {
		t.Errorf("expected status pending, got %q", got.Status)
	}
}

func TestTimeoutWorker_RunStopsOnCancel(t *testing.T) {
	repo := newMockRepo()

	worker := NewTimeoutWorker(repo, TimeoutWorkerConfig{Interval: 10 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Worker stopped as expected.
	case <-time.After(time.Second):
		t.Fatal("worker did not stop within timeout")
	}
}

func TestTimeoutWorker_MultipleExpiredRequests(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	pastTime := time.Now().Add(-1 * time.Minute)

	// Create 3 requests, expire 2.
	r1, _ := svc.CreateRequest(ctx, "ws-1", "a1", "q1", nil, "", 3600, "", "", "")
	r2, _ := svc.CreateRequest(ctx, "ws-1", "a1", "q2", nil, "", 3600, "", "", "")
	r3, _ := svc.CreateRequest(ctx, "ws-1", "a1", "q3", nil, "", 3600, "", "", "")

	repo.requests[r1.ID].ExpiresAt = &pastTime
	repo.requests[r2.ID].ExpiresAt = &pastTime
	// r3 keeps future expiry.

	worker := NewTimeoutWorker(repo, TimeoutWorkerConfig{Interval: time.Second})
	worker.sweep(ctx)

	if repo.requests[r1.ID].Status != models.HumanRequestStatusExpired {
		t.Errorf("r1: expected expired, got %q", repo.requests[r1.ID].Status)
	}
	if repo.requests[r2.ID].Status != models.HumanRequestStatusExpired {
		t.Errorf("r2: expected expired, got %q", repo.requests[r2.ID].Status)
	}
	if repo.requests[r3.ID].Status != models.HumanRequestStatusPending {
		t.Errorf("r3: expected pending, got %q", repo.requests[r3.ID].Status)
	}
}

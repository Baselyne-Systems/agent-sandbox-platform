package human

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// syncMockRepo wraps mockRepo with a mutex to support concurrent benchmark access.
type syncMockRepo struct {
	mu   sync.Mutex
	mock *mockRepo
}

func newSyncMockRepo() *syncMockRepo {
	return &syncMockRepo{mock: newMockRepo()}
}

func (s *syncMockRepo) CreateRequest(ctx context.Context, req *models.HumanRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateRequest(ctx, req)
}

func (s *syncMockRepo) GetRequest(ctx context.Context, tenantID, id string) (*models.HumanRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetRequest(ctx, tenantID, id)
}

func (s *syncMockRepo) RespondToRequest(ctx context.Context, tenantID, id, response, responderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.RespondToRequest(ctx, tenantID, id, response, responderID)
}

func (s *syncMockRepo) ListRequests(ctx context.Context, tenantID, workspaceID string, status models.HumanRequestStatus, afterID string, limit int) ([]models.HumanRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListRequests(ctx, tenantID, workspaceID, status, afterID, limit)
}

func (s *syncMockRepo) UpsertDeliveryChannel(ctx context.Context, cfg *models.DeliveryChannelConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpsertDeliveryChannel(ctx, cfg)
}

func (s *syncMockRepo) GetDeliveryChannel(ctx context.Context, tenantID, userID, channelType string) (*models.DeliveryChannelConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetDeliveryChannel(ctx, tenantID, userID, channelType)
}

func (s *syncMockRepo) UpsertTimeoutPolicy(ctx context.Context, policy *models.TimeoutPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpsertTimeoutPolicy(ctx, policy)
}

func (s *syncMockRepo) GetTimeoutPolicy(ctx context.Context, tenantID, scope, scopeID string) (*models.TimeoutPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetTimeoutPolicy(ctx, tenantID, scope, scopeID)
}

func (s *syncMockRepo) ListEnabledChannels(ctx context.Context, tenantID string) ([]models.DeliveryChannelConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListEnabledChannels(ctx, tenantID)
}

func (s *syncMockRepo) ExpirePendingRequests(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ExpirePendingRequests(ctx)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkCreateRequest_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCreateRequest_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			_, err := svc.CreateRequest(ctx, "tenant-1",
				fmt.Sprintf("ws-%d", n%10), "agent-1",
				"Approve action?", []string{"yes", "no"},
				"context data", 300,
				models.HumanRequestTypeApproval,
				models.HumanRequestUrgencyNormal, "")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetRequest_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetRequest_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	requestIDs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		req, err := svc.CreateRequest(ctx, "tenant-1",
			fmt.Sprintf("ws-%d", i%10), "agent-1",
			fmt.Sprintf("Question %d?", i), nil, "", 3600, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
		requestIDs[i] = req.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			id := requestIDs[idx%len(requestIDs)]
			idx++
			_, err := svc.GetRequest(ctx, "tenant-1", id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkRespondToRequest_Throughput
// ---------------------------------------------------------------------------

func BenchmarkRespondToRequest_Throughput(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		req, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
			"Approve?", []string{"yes", "no"}, "", 300, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		err = svc.RespondToRequest(ctx, "tenant-1", req.ID, "approved", "human-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkCreateRequest_AllUrgencies
// ---------------------------------------------------------------------------

func BenchmarkCreateRequest_AllUrgencies(b *testing.B) {
	urgencies := []struct {
		name    string
		urgency models.HumanRequestUrgency
	}{
		{"low", models.HumanRequestUrgencyLow},
		{"normal", models.HumanRequestUrgencyNormal},
		{"high", models.HumanRequestUrgencyHigh},
		{"critical", models.HumanRequestUrgencyCritical},
	}

	for _, tc := range urgencies {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for b.Loop() {
				_, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
					"Approve?", []string{"yes", "no"}, "ctx", 300,
					models.HumanRequestTypeApproval, tc.urgency, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkCreateRequest_AllTypes
// ---------------------------------------------------------------------------

func BenchmarkCreateRequest_AllTypes(b *testing.B) {
	types := []struct {
		name        string
		requestType models.HumanRequestType
	}{
		{"approval", models.HumanRequestTypeApproval},
		{"question", models.HumanRequestTypeQuestion},
		{"escalation", models.HumanRequestTypeEscalation},
	}

	for _, tc := range types {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for b.Loop() {
				_, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
					"Question?", []string{"yes", "no"}, "ctx", 300,
					tc.requestType, models.HumanRequestUrgencyNormal, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkCreateRequest_LargeOptions
// ---------------------------------------------------------------------------

func BenchmarkCreateRequest_LargeOptions(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	options := make([]string, 20)
	for i := range options {
		options[i] = fmt.Sprintf("Option %d: This is a moderately detailed option description", i)
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
			"Select an option:", options, "context with many options", 300,
			models.HumanRequestTypeQuestion, models.HumanRequestUrgencyNormal, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkCreateRequest_LargeContext
// ---------------------------------------------------------------------------

func BenchmarkCreateRequest_LargeContext(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// 4096-character context string.
	largeContext := strings.Repeat("A", 4096)

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
			"Approve?", []string{"yes", "no"}, largeContext, 300,
			models.HumanRequestTypeApproval, models.HumanRequestUrgencyNormal, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkListRequests_50K
// ---------------------------------------------------------------------------

func BenchmarkListRequests_50K(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 50000; i++ {
		_, err := svc.CreateRequest(ctx, "tenant-1",
			fmt.Sprintf("ws-%d", i%100), "agent-1",
			fmt.Sprintf("q%d", i), nil, "", 3600, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListRequests(ctx, "tenant-1", "", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkListRequests_WithWorkspaceFilter
// ---------------------------------------------------------------------------

func BenchmarkListRequests_WithWorkspaceFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 5000; i++ {
		_, err := svc.CreateRequest(ctx, "tenant-1",
			fmt.Sprintf("ws-%d", i%50), "agent-1",
			fmt.Sprintf("q%d", i), nil, "", 3600, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListRequests(ctx, "tenant-1", "ws-7", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkListRequests_WithStatusFilter
// ---------------------------------------------------------------------------

func BenchmarkListRequests_WithStatusFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 5000; i++ {
		req, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
			fmt.Sprintf("q%d", i), nil, "", 3600, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
		// Respond to every other request to create a mix of statuses.
		if i%2 == 0 {
			_ = svc.RespondToRequest(ctx, "tenant-1", req.ID, "ok", "human-1")
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListRequests(ctx, "tenant-1", "", models.HumanRequestStatusPending, 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkRequestLifecycle_CreateRespondGet
// ---------------------------------------------------------------------------

func BenchmarkRequestLifecycle_CreateRespondGet(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		svc := NewService(newMockRepo())

		req, err := svc.CreateRequest(ctx, "tenant-1", "ws-1", "agent-1",
			"Approve deployment?", []string{"approve", "deny"}, "Deploy v2.3.0", 600,
			models.HumanRequestTypeApproval, models.HumanRequestUrgencyHigh, "task-1")
		if err != nil {
			b.Fatal(err)
		}

		err = svc.RespondToRequest(ctx, "tenant-1", req.ID, "approve", "human-1")
		if err != nil {
			b.Fatal(err)
		}

		_, err = svc.GetRequest(ctx, "tenant-1", req.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkConfigureDeliveryChannel
// ---------------------------------------------------------------------------

func BenchmarkConfigureDeliveryChannel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.ConfigureDeliveryChannel(ctx, "tenant-1", "user-1", "slack",
			"https://hooks.slack.com/services/T00/B00/xxx")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 13. BenchmarkSetTimeoutPolicy
// ---------------------------------------------------------------------------

func BenchmarkSetTimeoutPolicy(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.SetTimeoutPolicy(ctx, "tenant-1", "agent", "agent-1", 600, "escalate",
			[]string{"admin@example.com", "oncall@example.com"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 14. BenchmarkMultiTenantRequestIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantRequestIsolation(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	tenants := []string{"tenant-1", "tenant-2", "tenant-3", "tenant-4", "tenant-5"}
	requestIDs := make(map[string][]string)
	for _, t := range tenants {
		ids := make([]string, 200)
		for i := 0; i < 200; i++ {
			req, err := svc.CreateRequest(ctx, t,
				fmt.Sprintf("ws-%d", i%10), "agent-1",
				fmt.Sprintf("Question %d?", i), nil, "", 3600, "", "", "")
			if err != nil {
				b.Fatal(err)
			}
			ids[i] = req.ID
		}
		requestIDs[t] = ids
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			t := tenants[idx%len(tenants)]
			ids := requestIDs[t]
			id := ids[idx%len(ids)]
			idx++
			_, err := svc.GetRequest(ctx, t, id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 15. BenchmarkMixedRequestWorkload
// ---------------------------------------------------------------------------

func BenchmarkMixedRequestWorkload(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	// Seed requests for reads and responds.
	requestIDs := make([]string, 500)
	for i := 0; i < 500; i++ {
		req, err := svc.CreateRequest(ctx, "tenant-1",
			fmt.Sprintf("ws-%d", i%10), "agent-1",
			fmt.Sprintf("q%d", i), []string{"yes", "no"}, "", 3600,
			models.HumanRequestTypeQuestion, models.HumanRequestUrgencyNormal, "")
		if err != nil {
			b.Fatal(err)
		}
		requestIDs[i] = req.ID
	}

	var createCounter atomic.Int64
	var respondIdx atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		readIdx := 0
		for pb.Next() {
			roll := rng.Intn(100)
			switch {
			case roll < 70:
				// 70% reads
				id := requestIDs[readIdx%len(requestIDs)]
				readIdx++
				_, err := svc.GetRequest(ctx, "tenant-1", id)
				if err != nil {
					b.Fatal(err)
				}
			case roll < 90:
				// 20% creates
				n := createCounter.Add(1)
				req, err := svc.CreateRequest(ctx, "tenant-1",
					fmt.Sprintf("ws-new-%d", n), "agent-1",
					fmt.Sprintf("new-q-%d", n), nil, "", 3600, "", "", "")
				if err != nil {
					b.Fatal(err)
				}
				// Make the new request available for future responds.
				_ = req
			default:
				// 10% responds
				ri := respondIdx.Add(1)
				id := requestIDs[int(ri)%len(requestIDs)]
				// Best-effort: may fail if already responded; that's fine for benchmarking throughput.
				_ = svc.RespondToRequest(ctx, "tenant-1", id, "ok", "human-1")
			}
		}
	})
}

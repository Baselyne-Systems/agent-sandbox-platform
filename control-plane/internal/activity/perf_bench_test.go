package activity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func (s *syncMockRepo) InsertAction(ctx context.Context, record *models.ActionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.InsertAction(ctx, record)
}

func (s *syncMockRepo) GetAction(ctx context.Context, tenantID, id string) (*models.ActionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetAction(ctx, tenantID, id)
}

func (s *syncMockRepo) QueryActions(ctx context.Context, tenantID string, filter QueryFilter) ([]models.ActionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.QueryActions(ctx, tenantID, filter)
}

func (s *syncMockRepo) QueryActionsAll(ctx context.Context, tenantID string, filter QueryFilter, batchSize int, fn func([]models.ActionRecord) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.QueryActionsAll(ctx, tenantID, filter, batchSize, fn)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func makePayload(size int) json.RawMessage {
	m := make(map[string]string)
	for i := 0; len(mustMarshal(m)) < size; i++ {
		m[fmt.Sprintf("k%d", i)] = strings.Repeat("v", 40)
	}
	return mustMarshal(m)
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func seedActions(svc *Service, n int) {
	ctx := context.Background()
	outcomes := []models.ActionOutcome{
		models.ActionOutcomeAllowed,
		models.ActionOutcomeDenied,
		models.ActionOutcomeEscalated,
		models.ActionOutcomeError,
	}
	for i := 0; i < n; i++ {
		rec := &models.ActionRecord{
			TenantID:    "tenant-1",
			WorkspaceID: fmt.Sprintf("ws-%d", i%10),
			AgentID:     fmt.Sprintf("agent-%d", i%5),
			TaskID:      fmt.Sprintf("task-%d", i%20),
			ToolName:    fmt.Sprintf("tool-%d", i%8),
			Outcome:     outcomes[i%len(outcomes)],
			Parameters:  json.RawMessage(`{"cmd":"ls"}`),
		}
		svc.RecordAction(ctx, rec)
	}
}

func seedActionsForTenant(svc *Service, tenantID string, n int) {
	ctx := context.Background()
	for i := 0; i < n; i++ {
		rec := &models.ActionRecord{
			TenantID:    tenantID,
			WorkspaceID: fmt.Sprintf("ws-%d", i%5),
			AgentID:     fmt.Sprintf("agent-%d", i%3),
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeAllowed,
			Parameters:  json.RawMessage(`{"cmd":"echo"}`),
		}
		svc.RecordAction(ctx, rec)
	}
}

// ---------------------------------------------------------------------------
// 1. BenchmarkRecordAction_Parallel
// ---------------------------------------------------------------------------

func BenchmarkRecordAction_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rec := &models.ActionRecord{
				WorkspaceID: "ws-1",
				AgentID:     "agent-1",
				ToolName:    "shell",
				Outcome:     models.ActionOutcomeAllowed,
				Parameters:  json.RawMessage(`{"cmd":"ls","args":["-la"]}`),
			}
			if _, err := svc.RecordAction(ctx, rec); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetAction_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetAction_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	ids := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		rec := &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeAllowed,
		}
		id, _ := svc.RecordAction(ctx, rec)
		ids[i] = id
	}

	var idx atomic.Int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := idx.Add(1) % int64(len(ids))
			if _, err := svc.GetAction(ctx, "tenant-1", ids[i]); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkRecordAction_VaryingPayloadSizes
// ---------------------------------------------------------------------------

func BenchmarkRecordAction_VaryingPayloadSizes(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100B", 100},
		{"1KB", 1024},
		{"4KB", 4096},
		{"16KB", 16384},
		{"64KB", 65536},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()
			payload := makePayload(sz.size)

			for b.Loop() {
				rec := &models.ActionRecord{
					WorkspaceID: "ws-1",
					AgentID:     "agent-1",
					ToolName:    "http_request",
					Outcome:     models.ActionOutcomeAllowed,
					Parameters:  payload,
					Result:      payload,
				}
				if _, err := svc.RecordAction(ctx, rec); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkRecordAction_AllOutcomes
// ---------------------------------------------------------------------------

func BenchmarkRecordAction_AllOutcomes(b *testing.B) {
	outcomes := []struct {
		name    string
		outcome models.ActionOutcome
	}{
		{"Allowed", models.ActionOutcomeAllowed},
		{"Denied", models.ActionOutcomeDenied},
		{"Escalated", models.ActionOutcomeEscalated},
		{"Error", models.ActionOutcomeError},
	}

	for _, o := range outcomes {
		b.Run(o.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for b.Loop() {
				rec := &models.ActionRecord{
					WorkspaceID:  "ws-1",
					AgentID:      "agent-1",
					ToolName:     "shell",
					Outcome:      o.outcome,
					DenialReason: "policy violation",
				}
				if _, err := svc.RecordAction(ctx, rec); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkRecordAction_WithLatencies
// ---------------------------------------------------------------------------

func BenchmarkRecordAction_WithLatencies(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()
	evalLat := int64(250)
	execLat := int64(5000)

	for b.Loop() {
		rec := &models.ActionRecord{
			WorkspaceID:         "ws-1",
			AgentID:             "agent-1",
			ToolName:            "shell",
			Outcome:             models.ActionOutcomeAllowed,
			Parameters:          json.RawMessage(`{"cmd":"ls"}`),
			Result:              json.RawMessage(`{"output":"file.txt"}`),
			EvaluationLatencyUs: &evalLat,
			ExecutionLatencyUs:  &execLat,
		}
		if _, err := svc.RecordAction(ctx, rec); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkQueryActions_ScalingExtended
// ---------------------------------------------------------------------------

func BenchmarkQueryActions_ScalingExtended(b *testing.B) {
	for _, n := range []int{100, 1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			seedActions(svc, n)

			ctx := context.Background()
			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{
					WorkspaceID: "ws-0",
					Limit:       50,
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkQueryActions_AllFilters
// ---------------------------------------------------------------------------

func BenchmarkQueryActions_AllFilters(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	seedActions(svc, 5000)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{
			WorkspaceID: "ws-0",
			AgentID:     "agent-0",
			ToolName:    "tool-0",
			Outcome:     models.ActionOutcomeAllowed,
			Limit:       50,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkQueryActions_TimeRange
// ---------------------------------------------------------------------------

func BenchmarkQueryActions_TimeRange(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	seedActions(svc, 5000)
	ctx := context.Background()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{
			StartTime: &start,
			EndTime:   &end,
			Limit:     50,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkConfigureAlert_AllConditions
// ---------------------------------------------------------------------------

func BenchmarkConfigureAlert_AllConditions(b *testing.B) {
	conditions := []struct {
		name          string
		conditionType models.AlertConditionType
	}{
		{"DenialRate", models.AlertConditionDenialRate},
		{"ErrorRate", models.AlertConditionErrorRate},
		{"ActionVelocity", models.AlertConditionActionVelocity},
		{"BudgetBreach", models.AlertConditionBudgetBreach},
		{"StuckAgent", models.AlertConditionStuckAgent},
	}

	for _, c := range conditions {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			svc.SetAlertRepository(newMockAlertRepo())
			ctx := context.Background()

			for b.Loop() {
				_, err := svc.ConfigureAlert(ctx, "tenant-1", "alert-"+c.name, c.conditionType, 0.5, "agent-1", "https://hooks.example.com/alert")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkAlertLifecycle_ConfigResolve
// ---------------------------------------------------------------------------

func BenchmarkAlertLifecycle_ConfigResolve(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	alertRepo := newMockAlertRepo()
	svc.SetAlertRepository(alertRepo)
	ctx := context.Background()

	for b.Loop() {
		// Configure an alert.
		cfg, err := svc.ConfigureAlert(ctx, "tenant-1", "lifecycle-test", models.AlertConditionDenialRate, 0.5, "agent-1", "")
		if err != nil {
			b.Fatal(err)
		}

		// Create an alert directly to simulate triggering.
		alert := &models.Alert{
			ConfigID:      cfg.ID,
			AgentID:       "agent-1",
			ConditionType: models.AlertConditionDenialRate,
			Message:       "test alert",
			TriggeredAt:   time.Now(),
		}
		alertRepo.CreateAlert(ctx, alert)

		// List alerts.
		_, _, err = svc.ListAlerts(ctx, "agent-1", true, 10, "")
		if err != nil {
			b.Fatal(err)
		}

		// Get specific alert.
		_, err = svc.GetAlert(ctx, alert.ID)
		if err != nil {
			b.Fatal(err)
		}

		// Resolve it.
		if err := svc.ResolveAlert(ctx, alert.ID); err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkExportActions_JSON
// ---------------------------------------------------------------------------

func BenchmarkExportActions_JSON(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	seedActions(svc, 1000)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		var buf bytes.Buffer
		err := svc.ExportActions(ctx, "tenant-1", QueryFilter{}, ExportFormatJSON, func(data []byte, count int, isLast bool) error {
			buf.Write(data)
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkExportActions_CSV
// ---------------------------------------------------------------------------

func BenchmarkExportActions_CSV(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	seedActions(svc, 1000)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		var buf bytes.Buffer
		err := svc.ExportActions(ctx, "tenant-1", QueryFilter{}, ExportFormatCSV, func(data []byte, count int, isLast bool) error {
			buf.Write(data)
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 13. BenchmarkMultiTenantActionIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantActionIsolation(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())

	tenants := []string{"tenant-1", "tenant-2", "tenant-3", "tenant-4", "tenant-5"}
	for _, t := range tenants {
		seedActionsForTenant(svc, t, 500)
	}

	ctx := context.Background()
	var idx atomic.Int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := idx.Add(1) % int64(len(tenants))
			_, _, err := svc.QueryActions(ctx, tenants[i], QueryFilter{
				WorkspaceID: "ws-0",
				Limit:       50,
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 14. BenchmarkRecordAction_HighThroughput
// ---------------------------------------------------------------------------

func BenchmarkRecordAction_HighThroughput(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	var ops atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := ops.Add(1)
			rec := &models.ActionRecord{
				WorkspaceID: fmt.Sprintf("ws-%d", n%20),
				AgentID:     fmt.Sprintf("agent-%d", n%10),
				ToolName:    "shell",
				Outcome:     models.ActionOutcomeAllowed,
				Parameters:  json.RawMessage(`{"cmd":"echo","args":["hello"]}`),
			}
			if _, err := svc.RecordAction(ctx, rec); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 15. BenchmarkQueryActions_PaginationWalk
// ---------------------------------------------------------------------------

func BenchmarkQueryActions_PaginationWalk(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	seedActions(svc, 5000)
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		afterID := ""
		for {
			records, nextToken, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{
				AfterID: afterID,
				Limit:   100,
			})
			if err != nil {
				b.Fatal(err)
			}
			if nextToken == "" || len(records) == 0 {
				break
			}
			afterID = nextToken
		}
	}
}

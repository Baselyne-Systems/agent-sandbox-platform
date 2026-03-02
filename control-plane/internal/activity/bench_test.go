package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func BenchmarkRecordAction(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		rec := &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeAllowed,
			Parameters:  json.RawMessage(`{"cmd":"ls","args":["-la","/tmp"]}`),
		}
		_, err := svc.RecordAction(ctx, rec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecordAction_LargePayload(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// ~4KB JSON payload simulating real tool parameters
	largeParams := make(map[string]any)
	for i := 0; i < 50; i++ {
		largeParams[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d_with_some_extra_padding_to_make_it_realistic", i)
	}
	paramsJSON, _ := json.Marshal(largeParams)

	for b.Loop() {
		rec := &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "http_request",
			Outcome:     models.ActionOutcomeAllowed,
			Parameters:  json.RawMessage(paramsJSON),
			Result:      json.RawMessage(paramsJSON),
		}
		_, err := svc.RecordAction(ctx, rec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAction(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	rec := &models.ActionRecord{
		WorkspaceID: "ws-1",
		AgentID:     "agent-1",
		ToolName:    "shell",
		Outcome:     models.ActionOutcomeAllowed,
	}
	id, _ := svc.RecordAction(ctx, rec)

	for b.Loop() {
		_, err := svc.GetAction(ctx, "tenant-1", id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryActions_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			repo := newMockRepo()
			svc := NewService(repo)
			ctx := context.Background()

			for i := 0; i < n; i++ {
				rec := &models.ActionRecord{
					WorkspaceID: fmt.Sprintf("ws-%d", i%10),
					AgentID:     fmt.Sprintf("agent-%d", i%5),
					ToolName:    "shell",
					Outcome:     models.ActionOutcomeAllowed,
				}
				svc.RecordAction(ctx, rec)
			}

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

func BenchmarkQueryActions_MultiFilter(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		rec := &models.ActionRecord{
			WorkspaceID: fmt.Sprintf("ws-%d", i%10),
			AgentID:     fmt.Sprintf("agent-%d", i%5),
			ToolName:    fmt.Sprintf("tool-%d", i%3),
			Outcome:     models.ActionOutcomeAllowed,
		}
		svc.RecordAction(ctx, rec)
	}

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

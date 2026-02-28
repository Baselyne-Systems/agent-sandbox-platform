package economics

import (
	"context"
	"testing"
)

func BenchmarkRecordUsage(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.RecordUsage(ctx, "agent-1", "ws-1", "compute", "seconds", 100, 0.50)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetBudget(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "agent-1", 1000, "USD", "", 0)

	for b.Loop() {
		_, err := svc.GetBudget(ctx, "agent-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetBudget(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.SetBudget(ctx, "agent-1", 1000, "USD", "", 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckBudget(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "agent-1", 1000, "USD", "", 0)

	for b.Loop() {
		_, err := svc.CheckBudget(ctx, "agent-1", 50)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecordUsage_WithBudgetUpdate(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "agent-1", 1000000, "USD", "", 0)

	for b.Loop() {
		_, err := svc.RecordUsage(ctx, "agent-1", "ws-1", "compute", "seconds", 100, 0.50)
		if err != nil {
			b.Fatal(err)
		}
	}
}

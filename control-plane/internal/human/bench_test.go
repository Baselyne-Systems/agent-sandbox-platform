package human

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkCreateRequest(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.CreateRequest(ctx, "test-tenant", "ws-1", "agent-1", "Approve?", []string{"yes", "no"}, "context", 300, "", "", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRespondToRequest(b *testing.B) {
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		req, _ := svc.CreateRequest(ctx, "test-tenant", "ws", "a", "q", nil, "", 300, "", "", "")
		b.StartTimer()

		err := svc.RespondToRequest(ctx, "test-tenant", req.ID, "approved", "human-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListRequests_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				svc.CreateRequest(ctx, "test-tenant", "ws-1", "a", fmt.Sprintf("q%d", i), nil, "", 300, "", "", "")
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.ListRequests(ctx, "test-tenant", "", "", 50, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

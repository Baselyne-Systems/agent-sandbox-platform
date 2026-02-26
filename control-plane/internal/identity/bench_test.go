package identity

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkGenerateToken(b *testing.B) {
	for b.Loop() {
		_, _, err := GenerateToken()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashToken(b *testing.B) {
	token := "a]b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1"
	for b.Loop() {
		HashToken(token)
	}
}

func BenchmarkRegisterAgent(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	labels := map[string]string{"env": "prod", "team": "platform"}

	for b.Loop() {
		_, err := svc.RegisterAgent(ctx, "bench-agent", "benchmark", "owner-1", labels)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAgent(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "bench-agent", "", "owner-1", nil)

	for b.Loop() {
		_, err := svc.GetAgent(ctx, agent.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMintCredential(b *testing.B) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "bench-agent", "", "owner-1", nil)
	scopes := []string{"read", "write", "admin"}

	for b.Loop() {
		_, _, err := svc.MintCredential(ctx, agent.ID, scopes, 3600)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListAgents_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			repo := newMockRepo()
			svc := NewService(repo)
			ctx := context.Background()

			for i := 0; i < n; i++ {
				svc.RegisterAgent(ctx, fmt.Sprintf("agent-%d", i), "", "owner-1", nil)
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.ListAgents(ctx, "", "", 50, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDeactivateAgent(b *testing.B) {
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		repo := newMockRepo()
		svc := NewService(repo)
		agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil)
		for j := 0; j < 10; j++ {
			svc.MintCredential(ctx, agent.ID, []string{"read"}, 3600)
		}
		b.StartTimer()

		svc.DeactivateAgent(ctx, agent.ID)
	}
}

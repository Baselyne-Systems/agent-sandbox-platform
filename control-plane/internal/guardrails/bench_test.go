package guardrails

import (
	"context"
	"fmt"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func BenchmarkCreateRule(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	labels := map[string]string{"env": "prod", "team": "security"}

	for b.Loop() {
		_, err := svc.CreateRule(ctx, "test-tenant", "bench-rule", "benchmark", models.RuleTypeToolFilter, "tool == 'exec'", models.RuleActionDeny, 10, labels, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetRule(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "test-tenant", "bench-rule", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})

	for b.Loop() {
		_, err := svc.GetRule(ctx, "test-tenant", rule.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListRules_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				svc.CreateRule(ctx, "test-tenant", fmt.Sprintf("rule-%d", i), "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.ListRules(ctx, "test-tenant", "", false, 50, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCompilePolicy(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ids := make([]string, 50)
	for i := range ids {
		r, _ := svc.CreateRule(ctx, "test-tenant", fmt.Sprintf("rule-%d", i), "",
			models.RuleTypeToolFilter, "c", models.RuleActionDeny, i, nil, models.RuleScope{})
		ids[i] = r.ID
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.CompilePolicy(ctx, "test-tenant", ids)
		if err != nil {
			b.Fatal(err)
		}
	}
}

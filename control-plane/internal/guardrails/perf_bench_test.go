package guardrails

import (
	"context"
	"fmt"
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

func (s *syncMockRepo) CreateRule(ctx context.Context, rule *models.GuardrailRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateRule(ctx, rule)
}

func (s *syncMockRepo) GetRule(ctx context.Context, tenantID, id string) (*models.GuardrailRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetRule(ctx, tenantID, id)
}

func (s *syncMockRepo) UpdateRule(ctx context.Context, tenantID string, rule *models.GuardrailRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateRule(ctx, tenantID, rule)
}

func (s *syncMockRepo) DeleteRule(ctx context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.DeleteRule(ctx, tenantID, id)
}

func (s *syncMockRepo) ListRules(ctx context.Context, tenantID string, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListRules(ctx, tenantID, ruleType, enabledOnly, afterID, limit)
}

func (s *syncMockRepo) CreateSet(ctx context.Context, set *models.GuardrailSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateSet(ctx, set)
}

func (s *syncMockRepo) GetSet(ctx context.Context, tenantID, id string) (*models.GuardrailSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetSet(ctx, tenantID, id)
}

func (s *syncMockRepo) GetSetByName(ctx context.Context, tenantID, name string) (*models.GuardrailSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetSetByName(ctx, tenantID, name)
}

func (s *syncMockRepo) UpdateSet(ctx context.Context, tenantID string, set *models.GuardrailSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateSet(ctx, tenantID, set)
}

func (s *syncMockRepo) DeleteSet(ctx context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.DeleteSet(ctx, tenantID, id)
}

func (s *syncMockRepo) ListSets(ctx context.Context, tenantID string, afterID string, limit int) ([]models.GuardrailSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListSets(ctx, tenantID, afterID, limit)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkCreateRule_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCreateRule_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			_, err := svc.CreateRule(ctx, "tenant-1",
				fmt.Sprintf("rule-%d", n), "parallel bench rule",
				models.RuleTypeToolFilter, "exec,shell",
				models.RuleActionDeny, 10, nil, models.RuleScope{})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetRule_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetRule_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	ruleIDs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		rule, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("rule-%d", i), "",
			models.RuleTypeToolFilter, "exec",
			models.RuleActionDeny, i, nil, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
		ruleIDs[i] = rule.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			id := ruleIDs[idx%len(ruleIDs)]
			idx++
			_, err := svc.GetRule(ctx, "tenant-1", id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkCompilePolicy_Scaling
// ---------------------------------------------------------------------------

func BenchmarkCompilePolicy_Scaling(b *testing.B) {
	for _, n := range []int{1, 10, 50, 100, 500} {
		b.Run(fmt.Sprintf("rules=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			ids := make([]string, n)
			for i := 0; i < n; i++ {
				rule, err := svc.CreateRule(ctx, "tenant-1",
					fmt.Sprintf("rule-%d", i), "desc",
					models.RuleTypeToolFilter, "exec,shell,read_file",
					models.RuleActionDeny, i, nil, models.RuleScope{})
				if err != nil {
					b.Fatal(err)
				}
				ids[i] = rule.ID
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.CompilePolicy(ctx, "tenant-1", ids)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkCompilePolicy_ComplexScopes
// ---------------------------------------------------------------------------

func BenchmarkCompilePolicy_ComplexScopes(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	scope := models.RuleScope{
		AgentIDs:            []string{"agent-1", "agent-2", "agent-3", "agent-4", "agent-5"},
		ToolNames:           []string{"exec", "shell", "read_file", "write_file", "deploy"},
		TrustLevels:         []string{"new", "established", "trusted"},
		DataClassifications: []string{"public", "internal", "confidential", "restricted"},
	}

	ids := make([]string, 50)
	for i := 0; i < 50; i++ {
		rule, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("scoped-rule-%d", i), "complex scope",
			models.RuleTypeToolFilter, "exec,shell",
			models.RuleActionDeny, i,
			map[string]string{"env": "prod", "team": "security"},
			scope)
		if err != nil {
			b.Fatal(err)
		}
		ids[i] = rule.ID
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.CompilePolicy(ctx, "tenant-1", ids)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkSimulatePolicy_SingleRule
// ---------------------------------------------------------------------------

func BenchmarkSimulatePolicy_SingleRule(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, err := svc.CreateRule(ctx, "tenant-1",
		"deny-exec", "Block exec calls",
		models.RuleTypeToolFilter, "exec,shell",
		models.RuleActionDeny, 10, nil, models.RuleScope{})
	if err != nil {
		b.Fatal(err)
	}

	ids := []string{rule.ID}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.SimulatePolicy(ctx, "tenant-1", ids, "exec", nil, "agent-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkSimulatePolicy_ManyRules
// ---------------------------------------------------------------------------

func BenchmarkSimulatePolicy_ManyRules(b *testing.B) {
	for _, n := range []int{10, 50, 100, 500} {
		b.Run(fmt.Sprintf("rules=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			ids := make([]string, n)
			for i := 0; i < n; i++ {
				// Only the last rule matches "target_tool".
				toolList := fmt.Sprintf("tool_%d", i)
				if i == n-1 {
					toolList = "target_tool"
				}
				rule, err := svc.CreateRule(ctx, "tenant-1",
					fmt.Sprintf("rule-%d", i), "",
					models.RuleTypeToolFilter, toolList,
					models.RuleActionDeny, i, nil, models.RuleScope{})
				if err != nil {
					b.Fatal(err)
				}
				ids[i] = rule.ID
			}

			b.ResetTimer()
			for b.Loop() {
				_, err := svc.SimulatePolicy(ctx, "tenant-1", ids, "target_tool", nil, "agent-1")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkSimulatePolicy_ParameterCheck
// ---------------------------------------------------------------------------

func BenchmarkSimulatePolicy_ParameterCheck(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, err := svc.CreateRule(ctx, "tenant-1",
		"block-shadow", "Block /etc/shadow access",
		models.RuleTypeParameterCheck, "path=/etc/shadow",
		models.RuleActionDeny, 5, nil, models.RuleScope{})
	if err != nil {
		b.Fatal(err)
	}

	ids := []string{rule.ID}
	params := map[string]string{"path": "/etc/shadow"}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.SimulatePolicy(ctx, "tenant-1", ids, "read_file", params, "agent-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkCreateRule_AllTypes
// ---------------------------------------------------------------------------

func BenchmarkCreateRule_AllTypes(b *testing.B) {
	ruleTypes := []struct {
		name     string
		ruleType models.RuleType
	}{
		{"tool_filter", models.RuleTypeToolFilter},
		{"parameter_check", models.RuleTypeParameterCheck},
		{"rate_limit", models.RuleTypeRateLimit},
		{"budget_limit", models.RuleTypeBudgetLimit},
	}

	for _, tc := range ruleTypes {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for b.Loop() {
				_, err := svc.CreateRule(ctx, "tenant-1",
					"bench-rule", "benchmark",
					tc.ruleType, "condition=value",
					models.RuleActionDeny, 10, nil, models.RuleScope{})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkCreateRule_AllActions
// ---------------------------------------------------------------------------

func BenchmarkCreateRule_AllActions(b *testing.B) {
	actions := []struct {
		name   string
		action models.RuleAction
	}{
		{"allow", models.RuleActionAllow},
		{"deny", models.RuleActionDeny},
		{"escalate", models.RuleActionEscalate},
		{"log", models.RuleActionLog},
	}

	for _, tc := range actions {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for b.Loop() {
				_, err := svc.CreateRule(ctx, "tenant-1",
					"bench-rule", "benchmark",
					models.RuleTypeToolFilter, "exec",
					tc.action, 10, nil, models.RuleScope{})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkListRules_WithTypeFilter
// ---------------------------------------------------------------------------

func BenchmarkListRules_WithTypeFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	types := []models.RuleType{
		models.RuleTypeToolFilter,
		models.RuleTypeParameterCheck,
		models.RuleTypeRateLimit,
		models.RuleTypeBudgetLimit,
	}
	for i := 0; i < 1000; i++ {
		_, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("rule-%d", i), "",
			types[i%len(types)], "cond",
			models.RuleActionDeny, i, nil, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListRules(ctx, "tenant-1", models.RuleTypeToolFilter, false, 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkListRules_EnabledOnlyFilter
// ---------------------------------------------------------------------------

func BenchmarkListRules_EnabledOnlyFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		rule, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("rule-%d", i), "",
			models.RuleTypeToolFilter, "cond",
			models.RuleActionDeny, i, nil, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
		// Disable every other rule.
		if i%2 == 0 {
			rule.Enabled = false
			_, err = svc.UpdateRule(ctx, "tenant-1", rule)
			if err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListRules(ctx, "tenant-1", "", true, 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkCreateGuardrailSet
// ---------------------------------------------------------------------------

func BenchmarkCreateGuardrailSet(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	ruleIDs := make([]string, 20)
	for i := 0; i < 20; i++ {
		rule, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("rule-%d", i), "",
			models.RuleTypeToolFilter, "exec",
			models.RuleActionDeny, i, nil, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
		ruleIDs[i] = rule.ID
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.CreateSet(ctx, "tenant-1", "bench-set", "Benchmark guardrail set",
			ruleIDs, map[string]string{"env": "prod"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 13. BenchmarkMultiTenantRuleAccess
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantRuleAccess(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	tenants := []string{"tenant-1", "tenant-2", "tenant-3"}
	ruleIDs := make(map[string][]string)
	for _, t := range tenants {
		ids := make([]string, 100)
		for i := 0; i < 100; i++ {
			rule, err := svc.CreateRule(ctx, t,
				fmt.Sprintf("rule-%d", i), "",
				models.RuleTypeToolFilter, "exec",
				models.RuleActionDeny, i, nil, models.RuleScope{})
			if err != nil {
				b.Fatal(err)
			}
			ids[i] = rule.ID
		}
		ruleIDs[t] = ids
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			t := tenants[idx%len(tenants)]
			ids := ruleIDs[t]
			id := ids[idx%len(ids)]
			idx++
			_, err := svc.GetRule(ctx, t, id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 14. BenchmarkUpdateRule_Parallel
// ---------------------------------------------------------------------------

func BenchmarkUpdateRule_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	rules := make([]*models.GuardrailRule, 100)
	for i := 0; i < 100; i++ {
		rule, err := svc.CreateRule(ctx, "tenant-1",
			fmt.Sprintf("rule-%d", i), "desc",
			models.RuleTypeToolFilter, "exec",
			models.RuleActionDeny, i, nil, models.RuleScope{})
		if err != nil {
			b.Fatal(err)
		}
		rules[i] = rule
	}

	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			r := rules[int(n)%len(rules)]
			updated := *r
			updated.Priority = int(n)
			_, err := svc.UpdateRule(ctx, "tenant-1", &updated)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 15. BenchmarkDeleteRule_Throughput
// ---------------------------------------------------------------------------

func BenchmarkDeleteRule_Throughput(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		// Create a batch of rules to delete.
		ids := make([]string, 10)
		for i := 0; i < 10; i++ {
			rule, err := svc.CreateRule(ctx, "tenant-1",
				fmt.Sprintf("del-rule-%d", i), "",
				models.RuleTypeToolFilter, "exec",
				models.RuleActionDeny, i, nil, models.RuleScope{})
			if err != nil {
				b.Fatal(err)
			}
			ids[i] = rule.ID
		}
		b.StartTimer()

		for _, id := range ids {
			err := svc.DeleteRule(ctx, "tenant-1", id)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

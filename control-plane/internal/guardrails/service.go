package guardrails

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrRuleNotFound = errors.New("rule not found")
	ErrSetNotFound  = errors.New("guardrail set not found")
	ErrInvalidInput = errors.New("invalid input")
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

var validRuleTypes = map[models.RuleType]bool{
	models.RuleTypeToolFilter:     true,
	models.RuleTypeParameterCheck: true,
	models.RuleTypeRateLimit:      true,
	models.RuleTypeBudgetLimit:    true,
}

var validRuleActions = map[models.RuleAction]bool{
	models.RuleActionAllow:    true,
	models.RuleActionDeny:     true,
	models.RuleActionEscalate: true,
	models.RuleActionLog:      true,
}

// Service implements guardrails business logic on top of a Repository.
type Service struct {
	repo      Repository
	considered *ConsideredEvaluator
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetConsideredEvaluator attaches the considered evaluation tier.
func (s *Service) SetConsideredEvaluator(ce *ConsideredEvaluator) {
	s.considered = ce
}

// GetBehaviorReport delegates to the considered evaluator.
func (s *Service) GetBehaviorReport(ctx context.Context, agentID string, windowStart, windowEnd time.Time) (*models.BehaviorReport, error) {
	if s.considered == nil {
		return nil, errors.New("considered evaluator not configured")
	}
	return s.considered.GenerateReport(ctx, agentID, windowStart, windowEnd)
}

func (s *Service) CreateRule(ctx context.Context, name, description string, ruleType models.RuleType, condition string, action models.RuleAction, priority int, labels map[string]string, scope models.RuleScope) (*models.GuardrailRule, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if condition == "" {
		return nil, ErrInvalidInput
	}
	if !validRuleTypes[ruleType] {
		return nil, ErrInvalidInput
	}
	if !validRuleActions[action] {
		return nil, ErrInvalidInput
	}
	if labels == nil {
		labels = map[string]string{}
	}

	rule := &models.GuardrailRule{
		Name:        name,
		Description: description,
		Type:        ruleType,
		Condition:   condition,
		Action:      action,
		Priority:    priority,
		Enabled:     true,
		Labels:      labels,
		Scope:       scope,
	}
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) GetRule(ctx context.Context, id string) (*models.GuardrailRule, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	rule, err := s.repo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, ErrRuleNotFound
	}
	return rule, nil
}

func (s *Service) UpdateRule(ctx context.Context, rule *models.GuardrailRule) (*models.GuardrailRule, error) {
	if rule.ID == "" {
		return nil, ErrInvalidInput
	}
	if rule.Name == "" {
		return nil, ErrInvalidInput
	}
	if rule.Condition == "" {
		return nil, ErrInvalidInput
	}
	if !validRuleTypes[rule.Type] {
		return nil, ErrInvalidInput
	}
	if !validRuleActions[rule.Action] {
		return nil, ErrInvalidInput
	}
	if rule.Labels == nil {
		rule.Labels = map[string]string{}
	}
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) DeleteRule(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.DeleteRule(ctx, id)
}

func (s *Service) ListRules(ctx context.Context, ruleType models.RuleType, enabledOnly bool, pageSize int, pageToken string) ([]models.GuardrailRule, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	afterID, err := decodePageToken(pageToken)
	if err != nil {
		return nil, "", ErrInvalidInput
	}

	rules, err := s.repo.ListRules(ctx, ruleType, enabledOnly, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(rules) > pageSize {
		rules = rules[:pageSize]
		nextToken = encodePageToken(rules[pageSize-1].ID)
	}

	return rules, nextToken, nil
}

// compiledRuleScope matches the Rust evaluator's RuleScope struct.
type compiledRuleScope struct {
	AgentIDs            []string `json:"agent_ids,omitempty"`
	ToolNames           []string `json:"tool_names,omitempty"`
	TrustLevels         []string `json:"trust_levels,omitempty"`
	DataClassifications []string `json:"data_classifications,omitempty"`
}

// compiledRule matches the Rust evaluator's CompiledRule struct for JSON deserialization.
type compiledRule struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	RuleType  string             `json:"rule_type"`
	Condition string             `json:"condition"`
	Action    string             `json:"action"`
	Priority  int                `json:"priority"`
	Enabled   bool               `json:"enabled"`
	Scope     *compiledRuleScope `json:"scope,omitempty"`
}

// compiledPolicy matches the Rust evaluator's CompiledPolicy struct.
type compiledPolicy struct {
	Rules []compiledRule `json:"rules"`
}

// CompilePolicy fetches each rule by ID and produces a structured CompiledPolicy
// that the Rust runtime evaluator can deserialize and evaluate.
func (s *Service) CompilePolicy(ctx context.Context, ruleIDs []string) ([]byte, int, error) {
	if len(ruleIDs) == 0 {
		return nil, 0, ErrInvalidInput
	}

	var rules []compiledRule
	for _, id := range ruleIDs {
		rule, err := s.repo.GetRule(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		if rule == nil {
			return nil, 0, ErrRuleNotFound
		}
		cr := compiledRule{
			ID:        rule.ID,
			Name:      rule.Name,
			RuleType:  string(rule.Type),
			Condition: rule.Condition,
			Action:    string(rule.Action),
			Priority:  rule.Priority,
			Enabled:   rule.Enabled,
		}
		// Include scope if any field is non-empty.
		s := rule.Scope
		if len(s.AgentIDs) > 0 || len(s.ToolNames) > 0 || len(s.TrustLevels) > 0 || len(s.DataClassifications) > 0 {
			cr.Scope = &compiledRuleScope{
				AgentIDs:            s.AgentIDs,
				ToolNames:           s.ToolNames,
				TrustLevels:         s.TrustLevels,
				DataClassifications: s.DataClassifications,
			}
		}
		rules = append(rules, cr)
	}

	policy := compiledPolicy{Rules: rules}
	compiled, err := json.Marshal(policy)
	if err != nil {
		return nil, 0, err
	}
	return compiled, len(rules), nil
}

// SimulationResult holds the outcome of a policy simulation.
type SimulationResult struct {
	Verdict         string
	MatchedRuleID   string
	MatchedRuleName string
	Reason          string
}

// SimulatePolicy dry-runs a set of rules against a sample tool call, returning
// the verdict that would be produced. Rules are evaluated in priority order
// (ascending — lower priority number = higher precedence).
func (s *Service) SimulatePolicy(ctx context.Context, ruleIDs []string, toolName string, parameters map[string]string, agentID string) (*SimulationResult, error) {
	if len(ruleIDs) == 0 {
		return nil, ErrInvalidInput
	}
	if toolName == "" {
		return nil, ErrInvalidInput
	}

	// Fetch all rules.
	var rules []models.GuardrailRule
	for _, id := range ruleIDs {
		rule, err := s.repo.GetRule(ctx, id)
		if err != nil {
			return nil, err
		}
		if rule == nil {
			return nil, ErrRuleNotFound
		}
		rules = append(rules, *rule)
	}

	// Sort by priority ascending (lower number = higher precedence).
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })

	// Evaluate each enabled rule.
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		matched := false
		switch rule.Type {
		case models.RuleTypeToolFilter:
			// Condition is a comma-separated list of tool names.
			tools := strings.Split(rule.Condition, ",")
			for _, t := range tools {
				if strings.TrimSpace(t) == toolName {
					matched = true
					break
				}
			}
		case models.RuleTypeParameterCheck:
			// Condition format: "field=value"
			parts := strings.SplitN(rule.Condition, "=", 2)
			if len(parts) == 2 {
				field := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if parameters[field] == value {
					matched = true
				}
			}
		default:
			// rate_limit, budget_limit — not evaluated at request time.
			continue
		}

		if matched {
			verdict := "allow"
			reason := ""
			switch rule.Action {
			case models.RuleActionDeny:
				verdict = "deny"
				reason = "denied by rule: " + rule.Name
			case models.RuleActionEscalate:
				verdict = "escalate"
				reason = "escalated by rule: " + rule.Name
			case models.RuleActionAllow:
				verdict = "allow"
			case models.RuleActionLog:
				verdict = "allow" // log rules allow the action but log it
			}
			return &SimulationResult{
				Verdict:         verdict,
				MatchedRuleID:   rule.ID,
				MatchedRuleName: rule.Name,
				Reason:          reason,
			}, nil
		}
	}

	// No rule matched — default allow.
	return &SimulationResult{
		Verdict: "allow",
		Reason:  "no matching rule",
	}, nil
}

// --- GuardrailSet CRUD ---

func (s *Service) CreateSet(ctx context.Context, name, description string, ruleIDs []string, labels map[string]string) (*models.GuardrailSet, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if len(ruleIDs) == 0 {
		return nil, ErrInvalidInput
	}
	if labels == nil {
		labels = map[string]string{}
	}
	set := &models.GuardrailSet{
		Name:        name,
		Description: description,
		RuleIDs:     ruleIDs,
		Labels:      labels,
	}
	if err := s.repo.CreateSet(ctx, set); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *Service) GetSet(ctx context.Context, id string) (*models.GuardrailSet, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	set, err := s.repo.GetSet(ctx, id)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, ErrSetNotFound
	}
	return set, nil
}

func (s *Service) GetSetByName(ctx context.Context, name string) (*models.GuardrailSet, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	set, err := s.repo.GetSetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, ErrSetNotFound
	}
	return set, nil
}

func (s *Service) UpdateSet(ctx context.Context, set *models.GuardrailSet) (*models.GuardrailSet, error) {
	if set.ID == "" {
		return nil, ErrInvalidInput
	}
	if set.Name == "" {
		return nil, ErrInvalidInput
	}
	if len(set.RuleIDs) == 0 {
		return nil, ErrInvalidInput
	}
	if set.Labels == nil {
		set.Labels = map[string]string{}
	}
	if err := s.repo.UpdateSet(ctx, set); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *Service) DeleteSet(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.DeleteSet(ctx, id)
}

func (s *Service) ListSets(ctx context.Context, pageSize int, pageToken string) ([]models.GuardrailSet, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	afterID, err := decodePageToken(pageToken)
	if err != nil {
		return nil, "", ErrInvalidInput
	}

	sets, err := s.repo.ListSets(ctx, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(sets) > pageSize {
		sets = sets[:pageSize]
		nextToken = encodePageToken(sets[pageSize-1].ID)
	}

	return sets, nextToken, nil
}

// ResolveRuleIDs resolves a guardrail_policy_id string to a list of rule IDs.
// Supports two formats:
//   - "set:<name>" — looks up a named GuardrailSet and returns its rule IDs.
//   - "id1,id2,id3" — returns the comma-separated IDs directly (existing behavior).
func (s *Service) ResolveRuleIDs(ctx context.Context, policyID string) ([]string, error) {
	if strings.HasPrefix(policyID, "set:") {
		setName := strings.TrimPrefix(policyID, "set:")
		set, err := s.repo.GetSetByName(ctx, setName)
		if err != nil {
			return nil, err
		}
		if set == nil {
			return nil, ErrSetNotFound
		}
		return set.RuleIDs, nil
	}
	// Comma-separated rule IDs (existing format).
	ruleIDs := strings.Split(policyID, ",")
	for i := range ruleIDs {
		ruleIDs[i] = strings.TrimSpace(ruleIDs[i])
	}
	return ruleIDs, nil
}

func encodePageToken(id string) string {
	if id == "" {
		return ""
	}
	return id
}

func decodePageToken(token string) (string, error) {
	return token, nil
}

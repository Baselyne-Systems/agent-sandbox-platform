package guardrails

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrRuleNotFound = errors.New("rule not found")
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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRule(ctx context.Context, name, description string, ruleType models.RuleType, condition string, action models.RuleAction, priority int, labels map[string]string) (*models.GuardrailRule, error) {
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

// compiledRule matches the Rust evaluator's CompiledRule struct for JSON deserialization.
type compiledRule struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RuleType string `json:"rule_type"`
	Condition string `json:"condition"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
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
		rules = append(rules, compiledRule{
			ID:       rule.ID,
			Name:     rule.Name,
			RuleType: string(rule.Type),
			Condition: rule.Condition,
			Action:   string(rule.Action),
			Priority: rule.Priority,
			Enabled:  rule.Enabled,
		})
	}

	policy := compiledPolicy{Rules: rules}
	compiled, err := json.Marshal(policy)
	if err != nil {
		return nil, 0, err
	}
	return compiled, len(rules), nil
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

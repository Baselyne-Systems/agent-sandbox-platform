package identity

import (
	"context"
	"errors"
	"time"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrCredentialNotFound = errors.New("credential not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAgentInactive      = errors.New("agent is not active")
)

const (
	maxCredTTL     = 24 * time.Hour
	defaultPageSize = 50
	maxPageSize     = 100
)

// Service implements identity business logic on top of a Repository.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterAgent(ctx context.Context, name, description, ownerID string, labels map[string]string) (*models.Agent, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if ownerID == "" {
		return nil, ErrInvalidInput
	}
	if labels == nil {
		labels = map[string]string{}
	}
	agent := &models.Agent{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Status:      models.AgentStatusActive,
		Labels:      labels,
	}
	if err := s.repo.CreateAgent(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *Service) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	agent, err := s.repo.GetAgent(ctx, id)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

func (s *Service) ListAgents(ctx context.Context, ownerID string, status models.AgentStatus, pageSize int, pageToken string) ([]models.Agent, string, error) {
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

	// Fetch one extra to determine if there is a next page
	agents, err := s.repo.ListAgents(ctx, ownerID, status, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(agents) > pageSize {
		agents = agents[:pageSize]
		nextToken = encodePageToken(agents[pageSize-1].ID)
	}

	return agents, nextToken, nil
}

func (s *Service) DeactivateAgent(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.DeactivateAgent(ctx, id)
}

func (s *Service) MintCredential(ctx context.Context, agentID string, scopes []string, ttlSeconds int64) (*models.ScopedCredential, string, error) {
	if agentID == "" {
		return nil, "", ErrInvalidInput
	}
	if len(scopes) == 0 {
		return nil, "", ErrInvalidInput
	}
	if ttlSeconds <= 0 || time.Duration(ttlSeconds)*time.Second > maxCredTTL {
		return nil, "", ErrInvalidInput
	}

	agent, err := s.repo.GetAgent(ctx, agentID)
	if err != nil {
		return nil, "", err
	}
	if agent == nil {
		return nil, "", ErrAgentNotFound
	}
	if agent.Status != models.AgentStatusActive {
		return nil, "", ErrAgentInactive
	}

	rawToken, tokenHash, err := GenerateToken()
	if err != nil {
		return nil, "", err
	}

	cred := &models.ScopedCredential{
		AgentID:   agentID,
		Scopes:    scopes,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	if err := s.repo.CreateCredential(ctx, cred); err != nil {
		return nil, "", err
	}

	return cred, rawToken, nil
}

func (s *Service) RevokeCredential(ctx context.Context, credentialID string) error {
	if credentialID == "" {
		return ErrInvalidInput
	}
	return s.repo.RevokeCredential(ctx, credentialID)
}

// encodePageToken encodes a cursor ID as a page token.
func encodePageToken(id string) string {
	if id == "" {
		return ""
	}
	return id // UUIDs are already opaque; base64 wrapping adds no security here
}

// decodePageToken decodes a page token back to a cursor ID.
func decodePageToken(token string) (string, error) {
	return token, nil
}

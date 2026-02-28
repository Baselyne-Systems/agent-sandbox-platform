package human

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Deliverer sends a human interaction request notification to an external endpoint.
type Deliverer interface {
	Send(ctx context.Context, channel *models.DeliveryChannelConfig, request *models.HumanRequest) error
}

// WebhookPayload is the standard JSON payload delivered to all webhook endpoints.
// Community extensions (Slack bots, PagerDuty relays, etc.) consume this format.
type WebhookPayload struct {
	RequestID   string   `json:"request_id"`
	WorkspaceID string   `json:"workspace_id"`
	AgentID     string   `json:"agent_id"`
	TaskID      string   `json:"task_id,omitempty"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Context     string   `json:"context,omitempty"`
	Type        string   `json:"type"`
	Urgency     string   `json:"urgency"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
}

func newPayload(req *models.HumanRequest) WebhookPayload {
	p := WebhookPayload{
		RequestID:   req.ID,
		WorkspaceID: req.WorkspaceID,
		AgentID:     req.AgentID,
		TaskID:      req.TaskID,
		Question:    req.Question,
		Options:     req.Options,
		Context:     req.Context,
		Type:        string(req.Type),
		Urgency:     string(req.Urgency),
	}
	if req.ExpiresAt != nil {
		p.ExpiresAt = req.ExpiresAt.Format(time.RFC3339)
	}
	return p
}

// WebhookDeliverer sends notifications via HTTP POST with a standard JSON payload.
// All channel types (slack, email, teams, custom) receive the same payload format.
// Channel-specific formatting is handled by external adapters/extensions.
type WebhookDeliverer struct {
	Client *http.Client
}

func NewWebhookDeliverer() *WebhookDeliverer {
	return &WebhookDeliverer{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *WebhookDeliverer) Send(ctx context.Context, channel *models.DeliveryChannelConfig, request *models.HumanRequest) error {
	payload := newPayload(request)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, channel.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Bulkhead-Channel-Type", channel.ChannelType)
	httpReq.Header.Set("X-Bulkhead-Request-ID", request.ID)

	resp, err := d.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("webhook POST to %s: %w", channel.Endpoint, err)
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// notifyChannels delivers a newly created request to all enabled channels.
// Delivery is best-effort — failures are logged but do not fail the request.
func (s *Service) notifyChannels(ctx context.Context, req *models.HumanRequest) {
	if s.deliverer == nil {
		return
	}

	channels, err := s.repo.ListEnabledChannels(ctx)
	if err != nil {
		log.Printf("delivery: failed to list channels: %v", err)
		return
	}

	for i := range channels {
		if err := s.deliverer.Send(ctx, &channels[i], req); err != nil {
			log.Printf("delivery: channel %s (%s) failed: %v", channels[i].ID, channels[i].ChannelType, err)
		}
	}
}

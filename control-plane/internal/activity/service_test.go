package activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	records map[string]*models.ActionRecord
	nextID  int
}

func newMockRepo() *mockRepo {
	return &mockRepo{records: make(map[string]*models.ActionRecord)}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) InsertAction(_ context.Context, record *models.ActionRecord) error {
	record.ID = m.nextUUID()
	record.RecordedAt = time.Now()
	cp := *record
	m.records[record.ID] = &cp
	return nil
}

func (m *mockRepo) GetAction(_ context.Context, id string) (*models.ActionRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	return &cp, nil
}

func (m *mockRepo) QueryActions(_ context.Context, filter QueryFilter) ([]models.ActionRecord, error) {
	var result []models.ActionRecord
	for _, r := range m.records {
		if filter.WorkspaceID != "" && r.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.AgentID != "" && r.AgentID != filter.AgentID {
			continue
		}
		if filter.TaskID != "" && r.TaskID != filter.TaskID {
			continue
		}
		if filter.ToolName != "" && r.ToolName != filter.ToolName {
			continue
		}
		if filter.Outcome != "" && r.Outcome != filter.Outcome {
			continue
		}
		if filter.AfterID != "" && r.ID <= filter.AfterID {
			continue
		}
		result = append(result, *r)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func validRecord() *models.ActionRecord {
	return &models.ActionRecord{
		WorkspaceID: "ws-1",
		AgentID:     "agent-1",
		ToolName:    "shell",
		Outcome:     models.ActionOutcomeAllowed,
		Parameters:  json.RawMessage(`{"cmd":"ls"}`),
	}
}

func TestRecordAction_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	rec := validRecord()
	id, err := svc.RecordAction(context.Background(), rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty record ID")
	}
}

func TestRecordAction_Validation(t *testing.T) {
	svc := NewService(newMockRepo())

	tests := []struct {
		name   string
		modify func(*models.ActionRecord)
	}{
		{"empty workspace_id", func(r *models.ActionRecord) { r.WorkspaceID = "" }},
		{"empty agent_id", func(r *models.ActionRecord) { r.AgentID = "" }},
		{"empty tool_name", func(r *models.ActionRecord) { r.ToolName = "" }},
		{"empty outcome", func(r *models.ActionRecord) { r.Outcome = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := validRecord()
			tt.modify(rec)
			_, err := svc.RecordAction(context.Background(), rec)
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got: %v", err)
			}
		})
	}
}

func TestGetAction_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	rec := validRecord()
	id, _ := svc.RecordAction(context.Background(), rec)

	got, err := svc.GetAction(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID %q, got %q", id, got.ID)
	}
}

func TestGetAction_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetAction(context.Background(), "nonexistent")
	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got: %v", err)
	}
}

func TestQueryActions_WithFilters(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Insert records with different workspaces
	for i := 0; i < 5; i++ {
		rec := validRecord()
		rec.WorkspaceID = fmt.Sprintf("ws-%d", i%2)
		svc.RecordAction(ctx, rec)
	}

	records, _, err := svc.QueryActions(ctx, QueryFilter{WorkspaceID: "ws-0", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records for ws-0, got %d", len(records))
	}
}

func TestRecordAction_PublishesToBroker(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	subID, ch := svc.Broker().Subscribe()
	defer svc.Broker().Unsubscribe(subID)

	rec := validRecord()
	id, err := svc.RecordAction(ctx, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case got := <-ch:
		if got.ID != id {
			t.Errorf("expected record ID %q, got %q", id, got.ID)
		}
		if got.ToolName != "shell" {
			t.Errorf("expected tool 'shell', got %q", got.ToolName)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broker event")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker()

	id1, ch1 := broker.Subscribe()
	id2, ch2 := broker.Subscribe()
	defer broker.Unsubscribe(id1)
	defer broker.Unsubscribe(id2)

	rec := &models.ActionRecord{
		ID:          "rec-1",
		WorkspaceID: "ws-1",
		AgentID:     "agent-1",
		ToolName:    "shell",
		Outcome:     models.ActionOutcomeAllowed,
	}
	broker.Publish(rec)

	select {
	case got := <-ch1:
		if got.ID != "rec-1" {
			t.Errorf("subscriber 1: expected ID 'rec-1', got %q", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber 1: timed out")
	}

	select {
	case got := <-ch2:
		if got.ID != "rec-1" {
			t.Errorf("subscriber 2: expected ID 'rec-1', got %q", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber 2: timed out")
	}
}

func TestBroker_UnsubscribeClosesChannel(t *testing.T) {
	broker := NewBroker()
	id, ch := broker.Subscribe()
	broker.Unsubscribe(id)

	_, ok := <-ch
	if ok {
		t.Error("expected channel to be closed after unsubscribe")
	}
}

func TestQueryActions_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		rec := validRecord()
		svc.RecordAction(ctx, rec)
	}

	records, nextToken, err := svc.QueryActions(ctx, QueryFilter{Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if nextToken == "" {
		t.Error("expected next page token")
	}

	records2, nextToken2, err := svc.QueryActions(ctx, QueryFilter{AfterID: nextToken, Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records2) != 2 {
		t.Fatalf("expected 2 records on page 2, got %d", len(records2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

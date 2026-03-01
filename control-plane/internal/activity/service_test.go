package activity

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
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

func (m *mockRepo) GetAction(_ context.Context, _, id string) (*models.ActionRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	return &cp, nil
}

func (m *mockRepo) QueryActions(_ context.Context, _ string, filter QueryFilter) ([]models.ActionRecord, error) {
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

func (m *mockRepo) QueryActionsAll(_ context.Context, tenantID string, filter QueryFilter, batchSize int, fn func([]models.ActionRecord) error) error {
	filter.Limit = batchSize
	for {
		records, err := m.QueryActions(context.Background(), tenantID, filter)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return nil
		}
		if err := fn(records); err != nil {
			return err
		}
		if len(records) < batchSize {
			return nil
		}
		filter.AfterID = records[len(records)-1].ID
	}
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

	got, err := svc.GetAction(context.Background(), "tenant-1", id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID %q, got %q", id, got.ID)
	}
}

func TestGetAction_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetAction(context.Background(), "tenant-1", "nonexistent")
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

	records, _, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{WorkspaceID: "ws-0", Limit: 10})
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

	records, nextToken, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if nextToken == "" {
		t.Error("expected next page token")
	}

	records2, nextToken2, err := svc.QueryActions(ctx, "tenant-1", QueryFilter{AfterID: nextToken, Limit: 3})
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

// --- Export tests ---

func insertRecords(t *testing.T, svc *Service, n int) {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < n; i++ {
		rec := validRecord()
		rec.ToolName = fmt.Sprintf("tool-%d", i)
		if _, err := svc.RecordAction(ctx, rec); err != nil {
			t.Fatalf("insert record %d: %v", i, err)
		}
	}
}

func TestExportActions_JSON(t *testing.T) {
	svc := NewService(newMockRepo())
	insertRecords(t, svc, 5)

	var chunks [][]byte
	err := svc.ExportActions(context.Background(), "tenant-1", QueryFilter{}, ExportFormatJSON, func(data []byte, count int, isLast bool) error {
		chunks = append(chunks, data)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks (data + final), got %d", len(chunks))
	}

	// Last chunk should be the isLast=true signal with no data.
	lastChunk := chunks[len(chunks)-1]
	if len(lastChunk) != 0 {
		t.Errorf("expected empty final chunk, got %d bytes", len(lastChunk))
	}

	// First chunk should be NDJSON with 5 lines.
	lines := strings.Split(strings.TrimSpace(string(chunks[0])), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 NDJSON lines, got %d", len(lines))
	}

	// Each line should be valid JSON.
	for i, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
		if _, ok := obj["record_id"]; !ok {
			t.Errorf("line %d missing record_id field", i)
		}
	}
}

func TestExportActions_CSV(t *testing.T) {
	svc := NewService(newMockRepo())
	insertRecords(t, svc, 3)

	var allData []byte
	err := svc.ExportActions(context.Background(), "tenant-1", QueryFilter{}, ExportFormatCSV, func(data []byte, count int, isLast bool) error {
		allData = append(allData, data...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(bytes.NewReader(allData))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	// 1 header + 3 data rows
	if len(rows) != 4 {
		t.Fatalf("expected 4 CSV rows (header + 3 data), got %d", len(rows))
	}

	// Verify header.
	if rows[0][0] != "record_id" {
		t.Errorf("expected first header column 'record_id', got %q", rows[0][0])
	}

	// Verify data rows have non-empty record_id.
	for i := 1; i < len(rows); i++ {
		if rows[i][0] == "" {
			t.Errorf("row %d has empty record_id", i)
		}
	}
}

func TestExportActions_EmptyResult(t *testing.T) {
	svc := NewService(newMockRepo())

	var calls int
	err := svc.ExportActions(context.Background(), "tenant-1", QueryFilter{}, ExportFormatJSON, func(data []byte, count int, isLast bool) error {
		calls++
		if !isLast {
			t.Error("expected isLast=true for empty result")
		}
		if count != 0 {
			t.Errorf("expected count=0, got %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 sendFn call, got %d", calls)
	}
}

func TestExportActions_InvalidFormat(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.ExportActions(context.Background(), "tenant-1", QueryFilter{}, "xml", func(data []byte, count int, isLast bool) error {
		t.Error("sendFn should not be called for invalid format")
		return nil
	})
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got: %v", err)
	}
}

func TestFormatJSON_Output(t *testing.T) {
	records := []models.ActionRecord{
		{ID: "r1", WorkspaceID: "ws-1", AgentID: "a1", ToolName: "shell", Outcome: models.ActionOutcomeAllowed, RecordedAt: time.Now()},
		{ID: "r2", WorkspaceID: "ws-1", AgentID: "a1", ToolName: "read", Outcome: models.ActionOutcomeDenied, RecordedAt: time.Now()},
	}
	data, err := FormatJSON(records)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var obj map[string]any
	json.Unmarshal([]byte(lines[0]), &obj)
	if obj["record_id"] != "r1" {
		t.Errorf("expected record_id 'r1', got %v", obj["record_id"])
	}
}

func TestFormatCSV_Output(t *testing.T) {
	records := []models.ActionRecord{
		{ID: "r1", WorkspaceID: "ws-1", AgentID: "a1", ToolName: "shell", Outcome: models.ActionOutcomeAllowed, RecordedAt: time.Now()},
	}
	data, err := FormatCSV(records)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reader := csv.NewReader(bytes.NewReader(data))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (no header), got %d", len(rows))
	}
	if rows[0][0] != "r1" {
		t.Errorf("expected record_id 'r1', got %q", rows[0][0])
	}
}

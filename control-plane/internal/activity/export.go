package activity

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// ExportFormat identifies the output format for activity export.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// csvHeader is the column order for CSV exports.
var csvHeader = []string{
	"record_id", "workspace_id", "agent_id", "task_id", "tool_name",
	"outcome", "guardrail_rule_id", "denial_reason",
	"evaluation_latency_us", "execution_latency_us", "recorded_at",
	"parameters", "result",
}

// FormatJSON encodes records as newline-delimited JSON (NDJSON).
func FormatJSON(records []models.ActionRecord) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := range records {
		if err := enc.Encode(newExportRecord(&records[i])); err != nil {
			return nil, fmt.Errorf("encode record %s: %w", records[i].ID, err)
		}
	}
	return buf.Bytes(), nil
}

// FormatCSV encodes records as CSV rows (without header).
func FormatCSV(records []models.ActionRecord) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for i := range records {
		row := recordToCSVRow(&records[i])
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("write record %s: %w", records[i].ID, err)
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// CSVHeader returns the CSV header row as bytes.
func CSVHeader() []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write(csvHeader) //nolint:errcheck
	w.Flush()
	return buf.Bytes()
}

// exportRecord is the JSON-serializable view of an ActionRecord for export.
type exportRecord struct {
	RecordID            string          `json:"record_id"`
	WorkspaceID         string          `json:"workspace_id"`
	AgentID             string          `json:"agent_id"`
	TaskID              string          `json:"task_id,omitempty"`
	ToolName            string          `json:"tool_name"`
	Outcome             string          `json:"outcome"`
	GuardrailRuleID     string          `json:"guardrail_rule_id,omitempty"`
	DenialReason        string          `json:"denial_reason,omitempty"`
	EvaluationLatencyUs *int64          `json:"evaluation_latency_us,omitempty"`
	ExecutionLatencyUs  *int64          `json:"execution_latency_us,omitempty"`
	RecordedAt          time.Time       `json:"recorded_at"`
	Parameters          json.RawMessage `json:"parameters,omitempty"`
	Result              json.RawMessage `json:"result,omitempty"`
}

func newExportRecord(r *models.ActionRecord) *exportRecord {
	return &exportRecord{
		RecordID:            r.ID,
		WorkspaceID:         r.WorkspaceID,
		AgentID:             r.AgentID,
		TaskID:              r.TaskID,
		ToolName:            r.ToolName,
		Outcome:             string(r.Outcome),
		GuardrailRuleID:     r.GuardrailRuleID,
		DenialReason:        r.DenialReason,
		EvaluationLatencyUs: r.EvaluationLatencyUs,
		ExecutionLatencyUs:  r.ExecutionLatencyUs,
		RecordedAt:          r.RecordedAt,
		Parameters:          r.Parameters,
		Result:              r.Result,
	}
}

func recordToCSVRow(r *models.ActionRecord) []string {
	evalLat := ""
	if r.EvaluationLatencyUs != nil {
		evalLat = fmt.Sprintf("%d", *r.EvaluationLatencyUs)
	}
	execLat := ""
	if r.ExecutionLatencyUs != nil {
		execLat = fmt.Sprintf("%d", *r.ExecutionLatencyUs)
	}
	return []string{
		r.ID,
		r.WorkspaceID,
		r.AgentID,
		r.TaskID,
		r.ToolName,
		string(r.Outcome),
		r.GuardrailRuleID,
		r.DenialReason,
		evalLat,
		execLat,
		r.RecordedAt.Format(time.RFC3339),
		string(r.Parameters),
		string(r.Result),
	}
}

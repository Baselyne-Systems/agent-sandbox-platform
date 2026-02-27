package governance

import (
	"context"
	"testing"

	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/governance/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService())
}

func TestHandler_ClassifyData_PublicContent(t *testing.T) {
	h := newTestHandler()
	resp, err := h.ClassifyData(context.Background(), &pb.ClassifyDataRequest{
		Content:     []byte("Hello world, this is a test document"),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Classification != pb.DataClassification_DATA_CLASSIFICATION_PUBLIC {
		t.Errorf("classification = %v, want PUBLIC", resp.Classification)
	}
	if len(resp.DetectedPatterns) != 0 {
		t.Errorf("detected patterns = %v, want none", resp.DetectedPatterns)
	}
}

func TestHandler_ClassifyData_WithSSN(t *testing.T) {
	h := newTestHandler()
	resp, err := h.ClassifyData(context.Background(), &pb.ClassifyDataRequest{
		Content:     []byte("Patient SSN: 123-45-6789"),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Classification != pb.DataClassification_DATA_CLASSIFICATION_RESTRICTED {
		t.Errorf("classification = %v, want RESTRICTED", resp.Classification)
	}
	found := false
	for _, p := range resp.DetectedPatterns {
		if p == "ssn" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'ssn' in detected patterns, got %v", resp.DetectedPatterns)
	}
}

func TestHandler_ClassifyData_WithEmail(t *testing.T) {
	h := newTestHandler()
	resp, err := h.ClassifyData(context.Background(), &pb.ClassifyDataRequest{
		Content:     []byte("Contact: user@example.com for details"),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Classification != pb.DataClassification_DATA_CLASSIFICATION_INTERNAL {
		t.Errorf("classification = %v, want INTERNAL", resp.Classification)
	}
}

func TestHandler_ClassifyData_WithAWSKey(t *testing.T) {
	h := newTestHandler()
	resp, err := h.ClassifyData(context.Background(), &pb.ClassifyDataRequest{
		Content:     []byte("Key: AKIAIOSFODNN7EXAMPLE"),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Classification != pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL {
		t.Errorf("classification = %v, want CONFIDENTIAL", resp.Classification)
	}
}

func TestHandler_ClassifyData_EmptyContent(t *testing.T) {
	h := newTestHandler()
	_, err := h.ClassifyData(context.Background(), &pb.ClassifyDataRequest{
		Content: nil,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_CheckPolicy_RestrictedDenied(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CheckPolicy(context.Background(), &pb.CheckPolicyRequest{
		AgentId:            "agent-1",
		Destination:        "external-api",
		DataClassification: pb.DataClassification_DATA_CLASSIFICATION_RESTRICTED,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Allowed {
		t.Error("expected restricted data to be denied")
	}
	if resp.Reason == "" {
		t.Error("expected denial reason")
	}
}

func TestHandler_CheckPolicy_ConfidentialApproved(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CheckPolicy(context.Background(), &pb.CheckPolicyRequest{
		AgentId:            "agent-1",
		Destination:        "secure-storage",
		DataClassification: pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected confidential data allowed to approved destination")
	}
}

func TestHandler_CheckPolicy_ConfidentialDenied(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CheckPolicy(context.Background(), &pb.CheckPolicyRequest{
		AgentId:            "agent-1",
		Destination:        "unknown-api",
		DataClassification: pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Allowed {
		t.Error("expected confidential data denied to unapproved destination")
	}
}

func TestHandler_CheckPolicy_PublicAllowed(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CheckPolicy(context.Background(), &pb.CheckPolicyRequest{
		AgentId:            "agent-1",
		Destination:        "any-destination",
		DataClassification: pb.DataClassification_DATA_CLASSIFICATION_PUBLIC,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected public data to be allowed")
	}
}

func TestHandler_CheckPolicy_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.CheckPolicy(context.Background(), &pb.CheckPolicyRequest{
		AgentId:     "",
		Destination: "dest",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_ClassificationConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.DataClassification
		model DataClassification
	}{
		{pb.DataClassification_DATA_CLASSIFICATION_PUBLIC, ClassificationPublic},
		{pb.DataClassification_DATA_CLASSIFICATION_INTERNAL, ClassificationInternal},
		{pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL, ClassificationConfidential},
		{pb.DataClassification_DATA_CLASSIFICATION_RESTRICTED, ClassificationRestricted},
	}
	for _, tt := range tests {
		got := protoToClassification(tt.proto)
		if got != tt.model {
			t.Errorf("protoToClassification(%v) = %d, want %d", tt.proto, got, tt.model)
		}
		back := classificationToProto(got)
		if back != tt.proto {
			t.Errorf("classificationToProto(%d) = %v, want %v", got, back, tt.proto)
		}
	}
}

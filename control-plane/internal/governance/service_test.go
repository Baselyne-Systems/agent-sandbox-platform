package governance

import (
	"errors"
	"testing"
)

func TestClassifyData_SSN(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("SSN: 123-45-6789"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationRestricted {
		t.Errorf("expected Restricted, got %d", cls)
	}
	if !contains(patterns, "ssn") {
		t.Errorf("expected 'ssn' in patterns, got %v", patterns)
	}
}

func TestClassifyData_CreditCard(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("Card: 4111-1111-1111-1111"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationRestricted {
		t.Errorf("expected Restricted, got %d", cls)
	}
	if !contains(patterns, "credit_card") {
		t.Errorf("expected 'credit_card' in patterns, got %v", patterns)
	}
}

func TestClassifyData_Email(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("Contact: user@example.com"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationInternal {
		t.Errorf("expected Internal, got %d", cls)
	}
	if !contains(patterns, "email") {
		t.Errorf("expected 'email' in patterns, got %v", patterns)
	}
}

func TestClassifyData_Phone(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("Call: (555) 123-4567"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationInternal {
		t.Errorf("expected Internal, got %d", cls)
	}
	if !contains(patterns, "phone") {
		t.Errorf("expected 'phone' in patterns, got %v", patterns)
	}
}

func TestClassifyData_MultiplePatterns(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("SSN: 123-45-6789, email: test@example.com"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationRestricted {
		t.Errorf("expected Restricted (highest), got %d", cls)
	}
	if len(patterns) < 2 {
		t.Errorf("expected at least 2 patterns, got %v", patterns)
	}
}

func TestClassifyData_NoMatch(t *testing.T) {
	svc := NewService()
	cls, patterns, err := svc.ClassifyData([]byte("Hello, world!"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cls != ClassificationPublic {
		t.Errorf("expected Public for no-match, got %d", cls)
	}
	if len(patterns) != 0 {
		t.Errorf("expected no patterns, got %v", patterns)
	}
}

func TestClassifyData_EmptyContent(t *testing.T) {
	svc := NewService()
	_, _, err := svc.ClassifyData(nil, "text/plain")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty content, got: %v", err)
	}

	_, _, err = svc.ClassifyData([]byte{}, "text/plain")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty content, got: %v", err)
	}
}

func TestCheckPolicy_RestrictedDenied(t *testing.T) {
	svc := NewService()
	allowed, _, err := svc.CheckPolicy("agent-1", "internal-api", ClassificationRestricted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected restricted data to be denied")
	}
}

func TestCheckPolicy_ConfidentialApproved(t *testing.T) {
	svc := NewService()
	allowed, _, err := svc.CheckPolicy("agent-1", "internal-api", ClassificationConfidential)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected confidential data to be allowed to approved destination")
	}
}

func TestCheckPolicy_ConfidentialUnapproved(t *testing.T) {
	svc := NewService()
	allowed, _, err := svc.CheckPolicy("agent-1", "external-api", ClassificationConfidential)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected confidential data to be denied to unapproved destination")
	}
}

func TestCheckPolicy_PublicAllowed(t *testing.T) {
	svc := NewService()
	allowed, _, err := svc.CheckPolicy("agent-1", "anywhere", ClassificationPublic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected public data to be allowed")
	}
}

func TestCheckPolicy_InternalAllowed(t *testing.T) {
	svc := NewService()
	allowed, _, err := svc.CheckPolicy("agent-1", "anywhere", ClassificationInternal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected internal data to be allowed")
	}
}

func TestCheckPolicy_EmptyFields(t *testing.T) {
	svc := NewService()

	_, _, err := svc.CheckPolicy("", "dest", ClassificationPublic)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}

	_, _, err = svc.CheckPolicy("agent-1", "", ClassificationPublic)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty destination, got: %v", err)
	}
}

// --- InspectEgress tests ---

func TestInspectEgress_RestrictedContentBlocked(t *testing.T) {
	svc := NewService()
	allowed, reason, classification, patterns, err := svc.InspectEgress("agent-1", "internal-api", []byte("SSN: 123-45-6789"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected restricted content to be blocked")
	}
	if classification != ClassificationRestricted {
		t.Errorf("expected restricted, got %d", classification)
	}
	if !contains(patterns, "ssn") {
		t.Errorf("expected ssn pattern, got %v", patterns)
	}
	if reason == "" {
		t.Error("expected reason")
	}
}

func TestInspectEgress_PublicContentAllowed(t *testing.T) {
	svc := NewService()
	allowed, _, classification, _, err := svc.InspectEgress("agent-1", "any-dest", []byte("Hello, world!"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected public content to be allowed")
	}
	if classification != ClassificationPublic {
		t.Errorf("expected public, got %d", classification)
	}
}

func TestInspectEgress_ConfidentialApprovedDest(t *testing.T) {
	svc := NewService()
	allowed, _, classification, patterns, err := svc.InspectEgress("agent-1", "secure-storage", []byte("AKIAIOSFODNN7EXAMPLE"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected confidential content allowed to approved destination")
	}
	if classification != ClassificationConfidential {
		t.Errorf("expected confidential, got %d", classification)
	}
	if !contains(patterns, "aws_key") {
		t.Errorf("expected aws_key pattern, got %v", patterns)
	}
}

func TestInspectEgress_ConfidentialUnapprovedDest(t *testing.T) {
	svc := NewService()
	allowed, _, _, _, err := svc.InspectEgress("agent-1", "external-api", []byte("AKIAIOSFODNN7EXAMPLE"), "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected confidential content denied to unapproved destination")
	}
}

func TestInspectEgress_EmptyFields(t *testing.T) {
	svc := NewService()
	_, _, _, _, err := svc.InspectEgress("", "dest", []byte("test"), "text/plain")
	if err == nil {
		t.Error("expected error for empty agent_id")
	}
	_, _, _, _, err = svc.InspectEgress("agent-1", "", []byte("test"), "text/plain")
	if err == nil {
		t.Error("expected error for empty destination")
	}
	_, _, _, _, err = svc.InspectEgress("agent-1", "dest", nil, "text/plain")
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

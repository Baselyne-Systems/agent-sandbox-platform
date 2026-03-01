package governance

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// padText returns clean text of approximately the given byte length.
func padText(size int) []byte {
	base := "The quick brown fox jumps over the lazy dog. "
	reps := (size / len(base)) + 1
	return []byte(strings.Repeat(base, reps)[:size])
}

// embedPattern returns text of at least the given byte length with the
// given pattern embedded near the middle, surrounded by spaces for word
// boundary matching.
func embedPattern(size int, pattern string) []byte {
	wrapped := " " + pattern + " "
	if size < len(wrapped) {
		size = len(wrapped)
	}
	half := size / 2
	suffixLen := size - half - len(wrapped)
	if suffixLen < 0 {
		suffixLen = 0
	}
	prefix := padText(half)
	suffix := padText(suffixLen)
	return append(append(prefix, []byte(wrapped)...), suffix...)
}

// samplePatterns maps pattern names to sample values that match the regex.
var samplePatterns = map[string]string{
	"ssn":         "123-45-6789",
	"credit_card": "4111-1111-1111-1111",
	"aws_key":     "AKIAIOSFODNN7EXAMPLE",
	"email":       "user@example.com",
	"phone":       "(555) 123-4567",
}

// ---------------------------------------------------------------------------
// 1. BenchmarkClassifyData_Parallel
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	content := []byte("Please check SSN: 123-45-6789 and email user@example.com for verification.")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := svc.ClassifyData(content, "text/plain")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkClassifyData_PayloadScaling
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_PayloadScaling(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"10B", 10},
		{"100B", 100},
		{"1KB", 1024},
		{"10KB", 10240},
		{"100KB", 102400},
		{"1MB", 1048576},
	}

	svc := NewService()
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			b.ReportAllocs()
			// Embed an SSN near the end to ensure pattern matching traverses the payload.
			content := embedPattern(sz.size, "SSN: 123-45-6789")

			for b.Loop() {
				_, _, err := svc.ClassifyData(content, "text/plain")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. BenchmarkClassifyData_AllPatterns
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_AllPatterns(b *testing.B) {
	svc := NewService()

	for name, sample := range samplePatterns {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			content := embedPattern(1024, sample)

			for b.Loop() {
				_, detected, err := svc.ClassifyData(content, "text/plain")
				if err != nil {
					b.Fatal(err)
				}
				if len(detected) == 0 {
					b.Fatal("expected at least one pattern detected")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkClassifyData_MultiplePatterns
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_MultiplePatterns(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	content := []byte(
		"SSN: 123-45-6789, Card: 4111-1111-1111-1111, " +
			"Email: user@example.com, Phone: (555) 123-4567, " +
			"Some filler text to add realistic length to the document.")

	for b.Loop() {
		_, detected, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
		if len(detected) < 4 {
			b.Fatalf("expected at least 4 patterns, got %d", len(detected))
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkClassifyData_NoPatterns
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_NoPatterns(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	content := padText(1024)

	for b.Loop() {
		cls, detected, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
		if cls != ClassificationPublic {
			b.Fatalf("expected Public, got %d", cls)
		}
		if len(detected) != 0 {
			b.Fatalf("expected no patterns, got %v", detected)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkClassifyData_DensePatterns
// ---------------------------------------------------------------------------

func BenchmarkClassifyData_DensePatterns(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()

	// Build ~10KB content with a pattern every ~100 chars.
	var sb strings.Builder
	patterns := []string{
		"123-45-6789",
		"user@example.com",
		"4111-1111-1111-1111",
		"(555) 123-4567",
	}
	for i := 0; sb.Len() < 10240; i++ {
		sb.WriteString(strings.Repeat("x", 80))
		sb.WriteString(" ")
		sb.WriteString(patterns[i%len(patterns)])
		sb.WriteString(" ")
	}
	content := []byte(sb.String())

	for b.Loop() {
		_, _, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkCheckPolicy_AllClassifications
// ---------------------------------------------------------------------------

func BenchmarkCheckPolicy_AllClassifications(b *testing.B) {
	classifications := []struct {
		name           string
		classification DataClassification
	}{
		{"Public", ClassificationPublic},
		{"Internal", ClassificationInternal},
		{"Confidential", ClassificationConfidential},
		{"Restricted", ClassificationRestricted},
	}

	svc := NewService()
	for _, c := range classifications {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _, err := svc.CheckPolicy("agent-1", "internal-api", c.classification)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkCheckPolicy_ApprovedDestinations
// ---------------------------------------------------------------------------

func BenchmarkCheckPolicy_ApprovedDestinations(b *testing.B) {
	destinations := []string{"internal-api", "secure-storage", "audit-log"}

	svc := NewService()
	for _, dest := range destinations {
		b.Run(dest, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				allowed, _, err := svc.CheckPolicy("agent-1", dest, ClassificationConfidential)
				if err != nil {
					b.Fatal(err)
				}
				if !allowed {
					b.Fatalf("expected allowed for approved destination %q", dest)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkCheckPolicy_DeniedDestinations
// ---------------------------------------------------------------------------

func BenchmarkCheckPolicy_DeniedDestinations(b *testing.B) {
	destinations := []string{"external-api", "public-bucket", "unknown-service"}

	svc := NewService()
	for _, dest := range destinations {
		b.Run(dest, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				allowed, _, err := svc.CheckPolicy("agent-1", dest, ClassificationConfidential)
				if err != nil {
					b.Fatal(err)
				}
				if allowed {
					b.Fatalf("expected denied for unapproved destination %q", dest)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkCheckPolicy_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCheckPolicy_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := svc.CheckPolicy("agent-1", "internal-api", ClassificationConfidential)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 11. BenchmarkInspectEgress_Parallel
// ---------------------------------------------------------------------------

func BenchmarkInspectEgress_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	content := []byte("Email: user@example.com, SSN: 123-45-6789")

	var idx atomic.Int64
	agents := []string{"agent-1", "agent-2", "agent-3", "agent-4", "agent-5"}
	destinations := []string{"internal-api", "secure-storage", "external-api"}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := idx.Add(1)
			agent := agents[n%int64(len(agents))]
			dest := destinations[n%int64(len(destinations))]
			_, _, _, _, err := svc.InspectEgress(agent, dest, content, "text/plain")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 12. BenchmarkInspectEgress_LargePayload
// ---------------------------------------------------------------------------

func BenchmarkInspectEgress_LargePayload(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	// 100KB payload with an SSN embedded deep inside.
	content := embedPattern(102400, "SSN: 987-65-4321")

	for b.Loop() {
		_, _, cls, _, err := svc.InspectEgress("agent-1", "internal-api", content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
		if cls != ClassificationRestricted {
			b.Fatalf("expected Restricted, got %d", cls)
		}
	}
}

// ---------------------------------------------------------------------------
// 13. BenchmarkInspectEgress_CleanPayload
// ---------------------------------------------------------------------------

func BenchmarkInspectEgress_CleanPayload(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	// No sensitive patterns, approved destination -- fast path.
	content := padText(4096)

	for b.Loop() {
		allowed, _, cls, detected, err := svc.InspectEgress("agent-1", "internal-api", content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
		if !allowed {
			b.Fatal("expected allowed for clean payload to approved destination")
		}
		if cls != ClassificationPublic {
			b.Fatalf("expected Public, got %d", cls)
		}
		if len(detected) != 0 {
			b.Fatalf("expected no patterns, got %v", detected)
		}
	}
}

// ---------------------------------------------------------------------------
// 14. BenchmarkInspectEgress_DirtyPayload
// ---------------------------------------------------------------------------

func BenchmarkInspectEgress_DirtyPayload(b *testing.B) {
	b.ReportAllocs()
	svc := NewService()
	// Restricted patterns + unapproved destination.
	content := []byte(fmt.Sprintf(
		"Report contains SSN: 123-45-6789 and credit card: 4111-1111-1111-1111. %s",
		strings.Repeat("Additional context data. ", 100),
	))

	for b.Loop() {
		allowed, reason, cls, detected, err := svc.InspectEgress("agent-1", "external-api", content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
		if allowed {
			b.Fatal("expected denied for dirty payload to unapproved destination")
		}
		if cls != ClassificationRestricted {
			b.Fatalf("expected Restricted, got %d", cls)
		}
		if len(detected) < 2 {
			b.Fatalf("expected at least 2 patterns, got %v", detected)
		}
		if reason == "" {
			b.Fatal("expected non-empty reason")
		}
	}
}

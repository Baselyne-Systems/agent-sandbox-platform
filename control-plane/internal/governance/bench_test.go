package governance

import (
	"strings"
	"testing"
)

func BenchmarkClassifyData_SmallPayload(b *testing.B) {
	svc := NewService()
	content := []byte("SSN: 123-45-6789")

	for b.Loop() {
		_, _, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClassifyData_LargePayload(b *testing.B) {
	svc := NewService()
	content := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 1000) + "SSN: 987-65-4321")

	for b.Loop() {
		_, _, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClassifyData_NoMatch(b *testing.B) {
	svc := NewService()
	content := []byte(strings.Repeat("Hello world. ", 100))

	for b.Loop() {
		_, _, err := svc.ClassifyData(content, "text/plain")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPolicy(b *testing.B) {
	svc := NewService()

	for b.Loop() {
		_, _, err := svc.CheckPolicy("agent-1", "internal-api", ClassificationConfidential)
		if err != nil {
			b.Fatal(err)
		}
	}
}

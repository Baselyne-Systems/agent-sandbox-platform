package identity

import (
	"testing"
)

func TestGenerateToken_Uniqueness(t *testing.T) {
	token1, hash1, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token2, hash2, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("two generated tokens should not be equal")
	}
	if hash1 == hash2 {
		t.Error("two generated hashes should not be equal")
	}
}

func TestGenerateToken_Length(t *testing.T) {
	token, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 32 bytes hex-encoded = 64 characters
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
	// SHA-256 hex = 64 characters
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "abc123"
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Error("HashToken should be deterministic")
	}
}

func TestGenerateToken_HashMatchesRehash(t *testing.T) {
	token, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if HashToken(token) != hash {
		t.Error("HashToken(rawToken) should equal the hash returned by GenerateToken")
	}
}

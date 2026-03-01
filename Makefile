.PHONY: proto build build-go build-rust build-bkctl test test-go test-rust test-integration test-e2e test-e2e-full test-e2e-all dev clean

# Generate protobuf code for Go (Rust is generated at build time via build.rs)
proto:
	cd proto && buf generate

# Build everything
build: build-go build-rust

build-go:
	cd control-plane && go build ./...

build-bkctl:
	cd cmd/bkctl && go build -ldflags "-X github.com/Baselyne-Systems/bulkhead/cmd/bkctl/internal/cli.version=$$(git describe --tags --always 2>/dev/null || echo dev) -X github.com/Baselyne-Systems/bulkhead/cmd/bkctl/internal/cli.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X github.com/Baselyne-Systems/bulkhead/cmd/bkctl/internal/cli.date=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ../../bkctl .

build-rust:
	cd runtime && cargo build

# Run all tests
test: test-go test-rust

test-go:
	cd control-plane && go test ./...

test-rust:
	cd runtime && cargo test

# Integration tests (requires Docker)
test-integration:
	cd control-plane && go test -tags integration -count=1 -v ./internal/...

test-integration-%:
	cd control-plane && go test -tags integration -count=1 -v ./internal/$*/...

# E2E tests — control-plane only (mock runtime, fast, requires Docker for PostgreSQL)
test-e2e:
	cd control-plane && go test -count=1 -v -run 'Test[^F]' ./e2e/...

# E2E tests — full-stack (real runtime binary, requires Docker + Rust toolchain)
test-e2e-full: build-rust
	cd control-plane && RUNTIME_BINARY=../runtime/target/release/runtime go test -count=1 -v -run 'TestFullStack' ./e2e/...

# E2E tests — all (control-plane + full-stack)
test-e2e-all: build-rust
	cd control-plane && RUNTIME_BINARY=../runtime/target/release/runtime go test -count=1 -v ./e2e/...

# Start local dev dependencies
dev:
	docker compose -f deploy/docker-compose.yml up -d

# Stop local dev dependencies
dev-down:
	docker compose -f deploy/docker-compose.yml down

# Format code
fmt:
	cd control-plane && gofmt -w .
	cd runtime && cargo fmt

# Lint
lint:
	cd proto && buf lint
	cd control-plane && go vet ./...
	cd runtime && cargo clippy

# Clean build artifacts
clean:
	cd control-plane && go clean ./...
	cd runtime && cargo clean

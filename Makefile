.PHONY: proto build build-go build-rust build-bkctl test test-go test-rust test-integration test-e2e test-e2e-full test-e2e-all dev clean bench bench-integration bench-report

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

build-rust-release:
	cd runtime && cargo build --release

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
test-e2e-full: build-rust-release
	cd control-plane && RUNTIME_BINARY=$(CURDIR)/runtime/target/release/host-agent go test -count=1 -v -run 'TestFullStack' ./e2e/...

# E2E tests — all (control-plane + full-stack)
test-e2e-all: build-rust-release
	cd control-plane && RUNTIME_BINARY=$(CURDIR)/runtime/target/release/host-agent go test -count=1 -v ./e2e/...

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

# Benchmarks — unit (no external deps)
bench:
	cd control-plane && go test -bench=. -benchmem -run=^$$ -count=1 ./internal/...

bench-%:
	cd control-plane && go test -bench=. -benchmem -run=^$$ -count=1 ./internal/$*/...

# Benchmarks — integration (requires Docker for PostgreSQL)
bench-integration:
	cd control-plane && go test -tags integration -bench=. -benchmem -run=^$$ -count=1 ./internal/...

bench-integration-%:
	cd control-plane && go test -tags integration -bench=. -benchmem -run=^$$ -count=1 ./internal/$*/...

# Benchmarks — JSON output for CI (parseable by benchmark-action)
bench-report:
	cd control-plane && go test -bench=. -benchmem -run=^$$ -count=5 -json ./internal/... > bench-results.json

# Clean build artifacts
clean:
	cd control-plane && go clean ./...
	cd runtime && cargo clean

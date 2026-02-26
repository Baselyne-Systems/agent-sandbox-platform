.PHONY: proto build build-go build-rust test test-go test-rust dev clean

# Generate protobuf code for Go (Rust is generated at build time via build.rs)
proto:
	cd proto && buf generate

# Build everything
build: build-go build-rust

build-go:
	cd control-plane && go build ./...

build-rust:
	cd runtime && cargo build

# Run all tests
test: test-go test-rust

test-go:
	cd control-plane && go test ./...

test-rust:
	cd runtime && cargo test

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

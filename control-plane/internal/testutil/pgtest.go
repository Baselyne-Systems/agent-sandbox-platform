//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB wraps a database connection and cleanup function for integration tests.
type TestDB struct {
	DB      *sql.DB
	Cleanup func()
}

// MustSetupTestDB starts a PostgreSQL 16 container, runs all migrations, and
// returns a TestDB. Panics on failure — intended for use in TestMain.
func MustSetupTestDB() *TestDB {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(fmt.Sprintf("start postgres container: %v", err))
	}

	host, err := container.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("get container host: %v", err))
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		panic(fmt.Sprintf("get container port: %v", err))
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("open db: %v", err))
	}
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("ping db: %v", err))
	}

	runMigrations(db)

	return &TestDB{
		DB: db,
		Cleanup: func() {
			db.Close()
			_ = container.Terminate(ctx)
		},
	}
}

// runMigrations locates the migrations directory relative to the source file
// and executes all .sql files in sorted order.
func runMigrations(db *sql.DB) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot determine source file path")
	}
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		panic(fmt.Sprintf("read migrations dir %s: %v", migrationsDir, err))
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			panic(fmt.Sprintf("read migration %s: %v", f, err))
		}
		if _, err := db.Exec(string(data)); err != nil {
			panic(fmt.Sprintf("exec migration %s: %v", f, err))
		}
	}
}

// TruncateAll removes all rows from every table, respecting FK constraints.
func TruncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE
		workspace_snapshots, delivery_channels, timeout_policies,
		action_records, human_requests, scoped_credentials, tasks,
		agents, guardrail_rules, usage_records, budgets,
		workspaces, hosts
		CASCADE`)
	if err != nil {
		t.Fatalf("truncate all: %v", err)
	}
}

// SeedAgent inserts a minimal agent row and returns its UUID.
// Required for tables with FK references to agents (action_records, human_requests).
func SeedAgent(t *testing.T, db *sql.DB, name string) string {
	t.Helper()
	var id string
	err := db.QueryRow(
		`INSERT INTO agents (name, owner_id, status, labels)
		 VALUES ($1, 'test-owner', 'active', '{}')
		 RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed agent %q: %v", name, err)
	}
	return id
}

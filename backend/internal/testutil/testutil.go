package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	// In CI, use the pre-provisioned PostgreSQL service via DATABASE_URL.
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return setupWithExistingDB(t, ctx, dsn)
	}

	// Locally, spin up a testcontainer.
	return setupWithTestcontainer(t, ctx)
}

func setupWithExistingDB(t *testing.T, ctx context.Context, dsn string) *pgxpool.Pool {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	migrationsPath = filepath.ToSlash(migrationsPath)

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Ignore — migrations may already be applied by another test package
	}
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		t.Fatalf("failed to close migrate source: %v", srcErr)
	}
	if dbErr != nil {
		t.Fatalf("failed to close migrate db: %v", dbErr)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Clean all tables before this test to ensure isolation.
	cleanTables(t, pool)

	t.Cleanup(func() { pool.Close() })

	return pool
}

func setupWithTestcontainer(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "clubepay_test",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() { container.Terminate(ctx) }) //nolint:errcheck

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/clubepay_test?sslmode=disable", host, port.Port())

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	migrationsPath = filepath.ToSlash(migrationsPath)

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	t.Cleanup(func() { pool.Close() })

	return pool
}

// cleanTables truncates all tables to ensure test isolation when sharing a database.
func cleanTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{"usages", "referrals", "subscriptions", "plans", "businesses", "password_resets", "users"}
	for _, table := range tables {
		if _, err := pool.Exec(context.Background(), "DELETE FROM "+table); err != nil {
			t.Fatalf("failed to clean table %s: %v", table, err)
		}
	}
}

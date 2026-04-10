package integration

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type PostgresSetup struct {
	Ctx        context.Context
	ConnString string
	Pool       *pgxpool.Pool
}

func SetupPostgresWithMigrations(t *testing.T) PostgresSetup {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(
		ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase("petcontrol_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres testcontainer; ensure Docker is running: %v", err)
	}
	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to build connection string: %v", err)
	}

	if err := waitForPostgres(ctx, connString); err != nil {
		t.Fatalf("postgres not ready: %v", err)
	}

	migrator, err := migrate.New("file://"+resolveMigrationsPath(t), connString)
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}
	t.Cleanup(func() {
		if cerr, derr := migrator.Close(); cerr != nil || derr != nil {
			t.Logf("warning: migrate close errors: source=%v db=%v", cerr, derr)
		}
	})

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping pool: %v", err)
	}

	t.Cleanup(pool.Close)

	return PostgresSetup{
		Ctx:        ctx,
		ConnString: connString,
		Pool:       pool,
	}
}

func resolveMigrationsPath(t *testing.T) string {
	t.Helper()

	repoRoot := resolveRepoRoot(t)
	return filepath.Join(repoRoot, "infra", "migrations")
}

func resolveRepoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))
}

func waitForPostgres(ctx context.Context, connString string) error {
	var lastErr error
	for i := 0; i < 60; i++ {
		pool, err := pgxpool.New(ctx, connString)
		if err == nil {
			err = pool.Ping(ctx)
			pool.Close()
		}
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = errors.New("unknown postgres readiness error")
	}

	return fmt.Errorf("postgres not ready: %w", lastErr)
}

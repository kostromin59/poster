package pgcontainer

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PGContainerConfig struct {
	Database string
	User     string
	Password string
}

type PGContainer struct {
	testcontainers.Container
	cfg PGContainerConfig
	t   *testing.T
}

func New(t *testing.T, cfg PGContainerConfig) *PGContainer {
	t.Helper()

	container, err := postgres.Run(t.Context(), "postgres:18-trixie",
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		)),
	)

	if err != nil {
		t.Fatalf("unable to create generic container: %q", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(t.Context())
	})

	pgc := &PGContainer{
		cfg:       cfg,
		Container: container,
		t:         t,
	}

	return pgc
}

func (pgc *PGContainer) DSN() string {
	p, err := pgc.MappedPort(pgc.t.Context(), "5432")
	if err != nil {
		pgc.t.Fatalf("unable to get mapped port: %q", err)
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable", "localhost", p.Port(), pgc.cfg.User, pgc.cfg.Password, pgc.cfg.Database)
}

func (pgc *PGContainer) Migrate(path string) {
	db, err := sql.Open("pgx", pgc.DSN())
	if err != nil {
		pgc.t.Fatalf("unable to connect to database: %q", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		pgc.t.Fatalf("unable to set dialect: %q", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		pgc.t.Fatal("unable to get current path")
	}

	currentDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(filepath.Dir(currentDir))
	migrationsPath := filepath.Join(projectRoot, path)

	if err := goose.Up(db, migrationsPath); err != nil {
		pgc.t.Fatalf("unable to apply migrations: %q", err)
	}
}

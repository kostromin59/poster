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

func New(t *testing.T, cfg PGContainerConfig) (*PGContainer, error) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:18-trixie",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       cfg.Database,
			"POSTGRES_USER":     cfg.User,
			"POSTGRES_PASSWORD": cfg.Password,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	cont, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            false,
	})
	if err != nil {
		t.Fatalf("unable to create generic container: %q", err)
	}
	t.Cleanup(func() {
		_ = cont.Terminate(t.Context())
	})

	pgc := &PGContainer{
		cfg:       cfg,
		Container: cont,
		t:         t,
	}

	return pgc, nil
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

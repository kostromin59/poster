package pgxrepository_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/internal/models"
	"github.com/kostromin59/poster/pkg/pgcontainer"
)

func TestTagFindAll(t *testing.T) {
	if testing.Short() {
		t.Skip("this test is integration")
	}

	pgc := pgcontainer.New(t, pgcontainer.PGContainerConfig{
		Database: "poster",
		User:     "poster",
		Password: "poster",
	})

	pgc.Migrate("migrations")

	pool, err := pgxpool.New(t.Context(), pgc.DSN())
	if err != nil {
		t.Fatalf("unable to create pool: %q", err)
	}

	tagRepo := pgxrepository.NewTag(pool)

	t.Run("not found error", func(t *testing.T) {
		_, err := tagRepo.FindAll(t.Context())
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, models.ErrTagNotFound) {
			t.Errorf("expected error %+v but got %+v", models.ErrTagNotFound, err)
		}
	})

	expectedTags := []models.Tag{"my tag 1", "my tag 2"}
	for _, tag := range expectedTags {
		if _, err := pool.Exec(t.Context(), `INSERT INTO tags (tag) VALUES ($1)`, tag); err != nil {
			t.Fatalf("unable to insert tags: %q", err)
		}
	}

	t.Run("successful", func(t *testing.T) {
		tags, err := tagRepo.FindAll(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		if !reflect.DeepEqual(tags, expectedTags) {
			t.Errorf("expected tags %+v but got %+v", expectedTags, tags)
		}
	})
}

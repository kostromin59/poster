package pgxrepository_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/internal/models"
	"github.com/kostromin59/poster/pkg/pgcontainer"
)

func TestSourceFindAll(t *testing.T) {
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

	sourceRepo := pgxrepository.NewSource(pool)

	expectedSources := []models.Source{models.SourceTG, models.SourceWebsite}

	t.Run("find initial sources", func(t *testing.T) {
		sources, err := sourceRepo.FindAll(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		if !reflect.DeepEqual(sources, expectedSources) {
			t.Errorf("expected sources %+v but got %+v", expectedSources, sources)
		}
	})
}

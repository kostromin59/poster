package pgxrepository_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/internal/models"
	"github.com/kostromin59/poster/pkg/pgcontainer"
)

func TestPostCreate(t *testing.T) {
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

	postRepo := pgxrepository.NewPost(pool)

	media := []models.Media{
		{Filetype: "jpeg", URI: "some path"},
		{Filetype: "mp4", URI: "some path 2"},
	}

	mediaIDs := make([]models.MediaID, len(media))
	for i, m := range media {
		mediaRow := pool.QueryRow(t.Context(), `INSERT INTO media (filetype, uri) VALUES ($1, $2) RETURNING id`, m.Filetype, m.URI)
		if err := mediaRow.Scan(&m.ID); err != nil {
			t.Fatalf("unable to insert media: %q", err)
		}

		media[i] = m
		mediaIDs[i] = m.ID
	}

	t.Run("successful", func(t *testing.T) {
		dto := models.CreatePostDTO{
			Title:       "my title",
			Content:     "my content",
			PublishDate: time.Now().Add(30 * time.Minute),
			Tags:        []models.Tag{"tag1", "tag2"},
			Sources:     []models.Source{models.SourceTG, models.SourceWebsite},
			Media:       mediaIDs,
		}

		post, err := postRepo.Create(t.Context(), dto)
		if err != nil {
			t.Fatalf("unable to create post: %q", err)
		}

		if post.ID == "" {
			t.Error("expected not empty post id")
		}

		if post.Title != dto.Title {
			t.Errorf("expected post title %q but got %q", dto.Title, post.Title)
		}

		if post.Content != dto.Content {
			t.Errorf("expected post content %q but got %q", dto.Content, post.Content)
		}

		expectedPublishDate := dto.PublishDate.Format("2006-01-02 15:05:05")
		publishDate := post.PublishDate.Format("2006-01-02 15:05:05")
		if expectedPublishDate != publishDate {
			t.Errorf("expected post publish date %q but got %q", expectedPublishDate, publishDate)
		}

		if !reflect.DeepEqual(dto.Tags, post.Tags) {
			t.Errorf("expected post tags %+v but got %+v", dto.Tags, post.Tags)
		}

		if !reflect.DeepEqual(media, post.Media) {
			t.Errorf("expected post media %+v but got %+v", media, post.Media)
		}
	})

	t.Run("the second post with the same sources and tags", func(t *testing.T) {
		dto := models.CreatePostDTO{
			Title:       "my title 2",
			Content:     "my content 2",
			PublishDate: time.Now().Add(30 * time.Minute),
			Tags:        []models.Tag{"tag1", "tag2"},
			Sources:     []models.Source{models.SourceWebsite, models.SourceTG},
			Media:       nil,
		}

		post, err := postRepo.Create(t.Context(), dto)
		if err != nil {
			t.Fatalf("unable to create post: %q", err)
		}

		if post.ID == "" {
			t.Error("expected not empty post id")
		}

		if post.Title != dto.Title {
			t.Errorf("expected post title %q but got %q", dto.Title, post.Title)
		}

		if post.Content != dto.Content {
			t.Errorf("expected post content %q but got %q", dto.Content, post.Content)
		}

		expectedPublishDate := dto.PublishDate.Format("2006-01-02 15:05:05")
		publishDate := post.PublishDate.Format("2006-01-02 15:05:05")
		if expectedPublishDate != publishDate {
			t.Errorf("expected post publish date %q but got %q", expectedPublishDate, publishDate)
		}

		if !reflect.DeepEqual(dto.Tags, post.Tags) {
			t.Errorf("expected post tags %+v but got %+v", dto.Tags, post.Tags)
		}

		mediaLen := len(post.Media)
		if mediaLen != 0 {
			t.Errorf("expected media len %d but got %d", 0, mediaLen)
		}

		countSources := 0
		sourcesRow := pool.QueryRow(t.Context(), `SELECT count(*) FROM sources`)
		if err := sourcesRow.Scan(&countSources); err != nil {
			t.Fatalf("unable to get count of sources: %q", err)
		}

		if countSources != 2 {
			t.Errorf("expected sources count %d but got %d", 2, countSources)
		}

		countTags := 0
		tagsRow := pool.QueryRow(t.Context(), `SELECT count(*) FROM tags`)
		if err := tagsRow.Scan(&countTags); err != nil {
			t.Fatalf("unable to get count of tags: %q", err)
		}

		if countTags != 2 {
			t.Errorf("expected tags count %d but got %d", 2, countTags)
		}
	})

	t.Run("create with the same media", func(t *testing.T) {
		dto := models.CreatePostDTO{
			Title:       "my title 3",
			Content:     "my content 3",
			PublishDate: time.Now().Add(30 * time.Minute),
			Tags:        []models.Tag{"tag1", "tag2"},
			Sources:     []models.Source{models.SourceWebsite, models.SourceTG},
			Media:       mediaIDs,
		}

		post, err := postRepo.Create(t.Context(), dto)
		if err != nil {
			t.Fatalf("unable to create post: %q", err)
		}

		if post.ID == "" {
			t.Error("expected not empty post id")
		}

		if post.Title != dto.Title {
			t.Errorf("expected post title %q but got %q", dto.Title, post.Title)
		}

		if post.Content != dto.Content {
			t.Errorf("expected post content %q but got %q", dto.Content, post.Content)
		}

		expectedPublishDate := dto.PublishDate.Format("2006-01-02 15:05:05")
		publishDate := post.PublishDate.Format("2006-01-02 15:05:05")
		if expectedPublishDate != publishDate {
			t.Errorf("expected post publish date %q but got %q", expectedPublishDate, publishDate)
		}

		if !reflect.DeepEqual(dto.Tags, post.Tags) {
			t.Errorf("expected post tags %+v but got %+v", dto.Tags, post.Tags)
		}

		if !reflect.DeepEqual(media, post.Media) {
			t.Errorf("expected post media %+v but got %+v", media, post.Media)
		}
	})
}

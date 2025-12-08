package pgxrepository_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/internal/models"
	"github.com/kostromin59/poster/pkg/pgcontainer"
)

func TestPostFindPublished(t *testing.T) {
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

	t.Run("not found error", func(t *testing.T) {
		_, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{}, 0, 20)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		if !errors.Is(err, models.ErrPostNotFound) {
			t.Errorf("expected error %+v but got %+v", models.ErrPostNotFound, err)
		}
	})

	tags := []models.Tag{"tag1", "tag2"}
	slices.Sort(tags)
	for _, tag := range tags {
		if _, err := pool.Exec(t.Context(), `INSERT INTO tags (tag) VALUES ($1)`, tag); err != nil {
			t.Fatalf("unable to insert tag: %q", err)
		}
	}

	sources := []models.Source{models.SourceWebsite, models.SourceTG}
	slices.Sort(sources)

	posts := []models.Post{
		{
			Title:       "title 1",
			Content:     "content 1",
			PublishDate: time.Now().Add(-1 * time.Hour).Truncate(time.Second),
			Tags:        tags,
			Sources:     []models.Source{sources[0]},
		},
		{
			Title:       "title 2",
			Content:     "content 2",
			PublishDate: time.Now().Add(-1 * 24 * time.Hour).Truncate(time.Second),
			Tags:        []models.Tag{tags[0]},
			Sources:     sources,
		},
		{
			Title:       "title 3",
			Content:     "content 3",
			PublishDate: time.Now().Add(-7 * 24 * time.Hour).Truncate(time.Second),
			Tags:        []models.Tag{tags[1]},
			Sources:     []models.Source{sources[1]},
		},
		{
			Title:       "non published",
			Content:     "non publised content",
			PublishDate: time.Now().Add(1 * time.Hour).Truncate(time.Second),
			Tags:        tags,
			Sources:     sources,
		},
	}

	for i, p := range posts {
		postRow := pool.QueryRow(t.Context(), `INSERT INTO posts (title, content, publish_date) VALUES ($1, $2, $3) RETURNING id`, p.Title, p.Content, p.PublishDate)
		if err := postRow.Scan(&p.ID); err != nil {
			t.Fatalf("unable to insert post: %q", err)
		}

		posts[i] = p

		for _, s := range p.Sources {
			if _, err := pool.Exec(t.Context(), `INSERT INTO posts_sources (source, post_id) VALUES ($1, $2)`, s, p.ID); err != nil {
				t.Fatalf("unable to insert posts sources: %q", err)
			}
		}

		for _, tag := range p.Tags {
			if _, err := pool.Exec(t.Context(), `INSERT INTO posts_tags (tag, post_id) VALUES ($1, $2)`, tag, p.ID); err != nil {
				t.Fatalf("unable to insert posts tags: %q", err)
			}
		}
	}

	t.Run("find published without media", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, posts[:3], publishedPosts)
	})

	media := []models.Media{
		{Filetype: "jpeg", URI: "some path"},
		{Filetype: "mp4", URI: "some path 2"},
	}

	for i, m := range media {
		mediaRow := pool.QueryRow(t.Context(), `INSERT INTO media (filetype, uri) VALUES ($1, $2) RETURNING id`, m.Filetype, m.URI)
		if err := mediaRow.Scan(&m.ID); err != nil {
			t.Fatalf("unable to insert media: %q", err)
		}

		media[i] = m
	}

	posts = []models.Post{
		{
			ID:          posts[0].ID,
			Title:       "title 1",
			Content:     "content 1",
			PublishDate: time.Now().Add(-1 * time.Hour).Truncate(time.Second),
			Tags:        tags,
			Sources:     []models.Source{sources[0]},
			Media:       []models.Media{media[0]},
		},
		{
			ID:          posts[1].ID,
			Title:       "title 2",
			Content:     "content 2",
			PublishDate: time.Now().Add(-1 * 24 * time.Hour).Truncate(time.Second),
			Tags:        []models.Tag{tags[0]},
			Sources:     sources,
			Media:       []models.Media{media[1]},
		},
		{
			ID:          posts[2].ID,
			Title:       "title 3",
			Content:     "content 3",
			PublishDate: time.Now().Add(-7 * 24 * time.Hour).Truncate(time.Second),
			Tags:        []models.Tag{tags[1]},
			Sources:     []models.Source{sources[1]},
			Media:       media,
		},
		{
			ID:          posts[3].ID,
			Title:       "non published",
			Content:     "non publised content",
			PublishDate: time.Now().Add(1 * time.Hour).Truncate(time.Second),
			Tags:        tags,
			Sources:     sources,
			Media:       media,
		},
	}

	for _, p := range posts {
		for _, m := range p.Media {
			if _, err := pool.Exec(t.Context(), `INSERT INTO posts_media (media_id, post_id) VALUES ($1, $2)`, m.ID, p.ID); err != nil {
				t.Fatalf("unable to insert posts_media: %q", err)
			}
		}
	}

	t.Run("find publised with media", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, posts[:3], publishedPosts)
	})

	t.Run("find with title filter", func(t *testing.T) {
		title := "1"
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{Title: &title}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, []models.Post{posts[0]}, publishedPosts)
	})

	t.Run("find with published from filter", func(t *testing.T) {
		publishedFrom := time.Now().Add(-2 * 24 * time.Hour)
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{PublishedFrom: &publishedFrom}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, []models.Post{posts[0], posts[1]}, publishedPosts)
	})

	t.Run("find with tag filter", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{Tags: []string{string(tags[0])}}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, []models.Post{posts[0], posts[1]}, publishedPosts)
	})

	t.Run("find with many tags filter", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{Tags: []string{string(tags[0]), string(tags[1]), "unknown"}}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, posts[:3], publishedPosts)
	})

	t.Run("find with source filter", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{Sources: []string{string(sources[0]), "unknown"}}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, posts[:2], publishedPosts)
	})

	t.Run("find with many sources filter", func(t *testing.T) {
		publishedPosts, err := postRepo.FindPublished(t.Context(), models.PostSearchFilters{Sources: []string{string(sources[0]), string(sources[1])}}, 0, 20)
		if err != nil {
			t.Fatalf("unexpected error: %q", err)
		}

		comparePosts(t, posts[:3], publishedPosts)
	})
}

func comparePosts(t *testing.T, expected, got []models.Post) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("expected posts len %d but got %d", len(expected), len(got))
	}

	for i, expected := range expected {
		post := got[i]

		if post.ID != expected.ID {
			t.Errorf("expected post id %q but got %q", expected.ID, post.ID)
		}

		if post.Title != expected.Title {
			t.Errorf("expected post title %q but got %q", expected.Title, post.Title)
		}
		if post.Content != expected.Content {
			t.Errorf("expected post content %q but got %q", expected.Content, post.Content)
		}

		expectedPublishDate := expected.PublishDate.Format("2006-01-02 15:04:05")
		publishDate := post.PublishDate.Format("2006-01-02 15:04:05")

		if expectedPublishDate != publishDate {
			t.Errorf("expected post publishDate %q but got %q", expectedPublishDate, publishDate)
		}

		expectedTags := slices.Clone(expected.Tags)
		slices.Sort(expectedTags)

		tags := slices.Clone(post.Tags)
		slices.Sort(tags)

		if !slices.Equal(tags, expectedTags) {
			t.Errorf("expected post tags %+v but got %+v", expectedTags, tags)
		}

		expectedSources := slices.Clone(expected.Sources)
		slices.Sort(expectedSources)

		sources := slices.Clone(post.Sources)
		slices.Sort(sources)

		if !slices.Equal(sources, expectedSources) {
			t.Errorf("expected post sources %+v but got %+v", expectedSources, sources)
		}

		expectedMedia := slices.Clone(expected.Media)
		media := slices.Clone(post.Media)

		if !slices.Equal(media, expectedMedia) {
			t.Errorf("expected post media %+v but got %+v", expectedMedia, media)
		}
	}
}

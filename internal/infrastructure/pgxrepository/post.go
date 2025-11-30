package pgxrepository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/models"
)

type Post struct {
	pool *pgxpool.Pool
}

func NewPost(pool *pgxpool.Pool) *Post {
	return &Post{
		pool: pool,
	}
}

func (p *Post) Create(ctx context.Context, dto models.CreatePostDTO) (models.Post, error) {
	const op = "pgxrepository.Post.Create"

	log := slog.With(slog.String("op", op))

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return models.Post{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error("unable to rollback tx", slog.String("err", err.Error()))
		}
	}()

	for _, s := range dto.Sources {
		if _, err := tx.Exec(ctx, `INSERT INTO sources (source) VALUES ($1) ON CONFLICT (source) DO NOTHING`, s); err != nil {
			return models.Post{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	for _, t := range dto.Tags {
		if _, err := tx.Exec(ctx, `INSERT INTO tags (tag) VALUES ($1) ON CONFLICT (tag) DO NOTHING`, t); err != nil {
			return models.Post{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	post := models.Post{
		Title:       dto.Title,
		Content:     dto.Content,
		PublishDate: dto.PublishDate,
		Sources:     dto.Sources,
		Tags:        dto.Tags,
	}

	postRow := tx.QueryRow(ctx, `INSERT INTO posts (title, content, publish_date) VALUES ($1, $2, $3) RETURNING id`, dto.Title, dto.Content, dto.PublishDate)
	if err := postRow.Scan(&post.ID); err != nil {
		return models.Post{}, fmt.Errorf("%s: %w", op, err)
	}

	for _, s := range dto.Sources {
		if _, err := tx.Exec(ctx, `INSERT INTO posts_sources (source, post_id) VALUES ($1, $2)`, s, post.ID); err != nil {
			return models.Post{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	for _, t := range dto.Tags {
		if _, err := tx.Exec(ctx, `INSERT INTO posts_tags (tag, post_id) VALUES ($1, $2)`, t, post.ID); err != nil {
			return models.Post{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	postMedia := make([]models.Media, 0, len(dto.Media))
	for _, m := range dto.Media {
		media := models.Media{
			ID: m,
		}

		mediaRow := tx.QueryRow(ctx, `WITH inserted_media AS (
				INSERT INTO posts_media (media_id, post_id) 
				VALUES ($1, $2) 
				RETURNING media_id
			)
			SELECT 
				m.filetype,
				m.uri
			FROM inserted_media im
			LEFT JOIN media m ON m.id = im.media_id`, m, post.ID)
		if err := mediaRow.Scan(&media.Filetype, &media.URI); err != nil {
			return models.Post{}, fmt.Errorf("%s: %w", op, err)
		}

		postMedia = append(postMedia, media)
	}

	post.Media = postMedia

	if err := tx.Commit(ctx); err != nil {
		return models.Post{}, fmt.Errorf("%s: %w", op, err)
	}

	return post, nil
}

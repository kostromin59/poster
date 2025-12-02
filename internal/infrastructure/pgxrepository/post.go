package pgxrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/models"
)

type DBPost struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Content     string    `db:"content"`
	PublishDate time.Time `db:"publish_date"`
	Tags        []string  `db:"tags"`
	Sources     []string  `db:"sources"`
	Media       []byte    `db:"media"`
}

type DBPostMedia struct {
	ID       string `json:"id,omitempty"`
	Filetype string `json:"filetype,omitempty"`
	URI      string `json:"uri,omitempty"`
}

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

func (p *Post) FindPublished(ctx context.Context, filters models.PostSearchFilters, offset, limit uint64) ([]models.Post, error) {
	const op = "pgxrepository.Post.FindPublished"

	log := slog.With(slog.String("op", op))

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error("unable to rollback tx", slog.String("err", err.Error()))
		}
	}()

	query := squirrel.Select(
		"p.id",
		"p.title",
		"p.content",
		"p.publish_date",
		"array_agg(DISTINCT t.tag ORDER BY t.tag) AS tags",
		"array_agg(DISTINCT s.source ORDER BY s.source) AS sources",
		`(
			SELECT COALESCE(
					json_agg(
							json_build_object(
									'id', m.id,
									'filetype', m.filetype,
									'uri', m.uri
							)
					),
					'[]'::json
			)
			FROM posts_media pm
			LEFT JOIN media m ON m.id = pm.media_id
			WHERE pm.post_id = p.id
		) AS media`,
	).From("posts p").
		LeftJoin("posts_tags t ON t.post_id = p.id").
		LeftJoin("posts_sources s ON s.post_id = p.id").
		Where("p.publish_date <= NOW()").
		GroupBy("p.id", "p.title", "p.content", "p.publish_date").
		OrderBy("p.publish_date DESC").
		Offset(offset).
		Limit(limit)

	if filters.Title != nil {
		query = query.Where("p.title ILIKE ?", "%"+*filters.Title+"%")
	}

	if filters.PublishedFrom != nil {
		query = query.Where("p.publish_date >= ?", filters.PublishedFrom)
	}

	if len(filters.Tags) != 0 {
		query = query.Where(squirrel.Expr(
			"EXISTS (SELECT 1 FROM posts_tags pt WHERE pt.post_id = p.id AND pt.tag = ANY(?))",
			filters.Tags,
		))
	}

	if len(filters.Sources) != 0 {
		query = query.Where(squirrel.Expr(
			"EXISTS (SELECT 1 FROM posts_sources ps WHERE ps.post_id = p.id AND ps.source = ANY(?))",
			filters.Sources,
		))
	}

	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	dbPosts, err := pgx.CollectRows(rows, pgx.RowToStructByName[DBPost])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dbPosts) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrPostNotFound)
	}

	posts := make([]models.Post, len(dbPosts))
	for i, dbp := range dbPosts {
		var dbPostMedia []DBPostMedia
		if err := json.Unmarshal(dbp.Media, &dbPostMedia); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		postMedia := make([]models.Media, len(dbPostMedia))
		for i, dbpm := range dbPostMedia {
			postMedia[i] = models.Media{
				ID:       models.MediaID(dbpm.ID),
				Filetype: dbpm.Filetype,
				URI:      dbpm.URI,
			}
		}

		postTags := make([]models.Tag, len(dbp.Tags))
		for i, t := range dbp.Tags {
			postTags[i] = models.Tag(t)
		}

		postSources := make([]models.Source, len(dbp.Sources))
		for i, t := range dbp.Sources {
			postSources[i] = models.Source(t)
		}

		posts[i] = models.Post{
			ID:          models.PostID(dbp.ID),
			Title:       dbp.Title,
			Content:     dbp.Content,
			PublishDate: dbp.PublishDate,
			Tags:        postTags,
			Sources:     postSources,
			Media:       postMedia,
		}
	}

	return posts, nil
}

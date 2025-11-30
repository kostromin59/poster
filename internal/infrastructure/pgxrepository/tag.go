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

type Tag struct {
	pool *pgxpool.Pool
}

func NewTag(pool *pgxpool.Pool) *Tag {
	return &Tag{
		pool: pool,
	}
}

func (t *Tag) FindAll(ctx context.Context) ([]models.Tag, error) {
	const op = "pgxrepository.Tag.FindAll"

	log := slog.With(slog.String("op", op))

	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{
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

	rows, err := t.pool.Query(ctx, `SELECT tag FROM tags`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowTo[models.Tag])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrTagNotFound)
	}

	return tags, nil
}

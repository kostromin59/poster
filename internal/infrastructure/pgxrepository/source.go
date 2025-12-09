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

type Source struct {
	pool *pgxpool.Pool
}

func NewSource(pool *pgxpool.Pool) *Source {
	return &Source{
		pool: pool,
	}
}

func (t *Source) FindAll(ctx context.Context) ([]models.Source, error) {
	const op = "pgxrepository.Source.FindAll"

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

	rows, err := t.pool.Query(ctx, `SELECT source FROM sources`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	dbSources, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(dbSources) == 0 {
		return nil, fmt.Errorf("%s: %w", op, models.ErrSourceNotFound)
	}

	sources := make([]models.Source, len(dbSources))
	for i, dbt := range dbSources {
		sources[i] = models.Source(dbt)
	}

	return sources, nil
}

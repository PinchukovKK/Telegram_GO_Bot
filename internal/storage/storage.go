package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"main/internal/model"
	"time"
)

type SourceStoragePostgres struct {
	db *sqlx.DB
}

type dbSource struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	FeedUrl   string    `db:"feed_url"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *SourceStoragePostgres) Sources(ctx context.Context) ([]model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var source []dbSource
	if err := conn.SelectContext(ctx, &source, "SELECT * FROM sources"); err != nil {
		return nil, err
	}

	return lo.Map(source, func(sources dbSource, _ int) model.Source { return model.Source(sources) }), nil
}

func (s *SourceStoragePostgres) SourcesByID(ctx context.Context, id int64) (model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return model.Source{}, err
	}
	defer conn.Close()

	var source dbSource
	if err := conn.GetContext(ctx, &source, `SELECT * FROM sources WHERE id = $1`, id); err != nil {
		return model.Source{}, nil
	}

	modelSource := model.Source{
		ID:        id,
		Name:      source.Name,
		FeedUrl:   source.FeedUrl,
		CreatedAt: source.CreatedAt,
	}

	return modelSource, nil
}

func (s *SourceStoragePostgres) Add(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var id int64

	row := conn.QueryRowxContext(
		ctx,
		`INSERT INTO sources (name, feed_url, create_at) VALUES ($1, $2, $3) RETURNING id`,
		source.Name,
		source.FeedUrl,
		source.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SourceStoragePostgres) Delete(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "DELETE FROM sources WHERE id = $1", id); err != nil {
		return err
	}

	return nil
}

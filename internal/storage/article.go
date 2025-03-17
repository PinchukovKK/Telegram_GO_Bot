package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"main/internal/model"
	"time"
)

type ArticletoragePostgres struct {
	db *sqlx.DB
}

func NewArticaleStorage(db *sqlx.DB) *ArticletoragePostgres {
	return &ArticletoragePostgres{db: db}
}

type dbArticale struct {
	ID        int64     `db:"id"`
	SourceID  int64     `db:"source_id"`
	Name      string    `db:"name"`
	FeedUrl   string    `db:"feed_url"`
	CreatedAt time.Time `db:"created_at"`
}

func (s *ArticletoragePostgres) Store(ctx context.Context, article model.Article) error {}

func (s *ArticletoragePostgres) AllNotPosted(ctx context.Context, since time.Time, limit uint64) error {
}

func (s *SourceStoragePostgres) MarkPosted(ctx context.Context, id int64) error {}

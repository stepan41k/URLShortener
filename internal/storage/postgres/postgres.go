package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	mu sync.Mutex
	pool *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.Connect(context.Background(), storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer pool.Close()

	_, err = pool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{}, nil
}
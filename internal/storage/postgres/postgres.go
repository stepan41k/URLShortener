package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stepan41k/FullRestAPI/internal/storage"
)

type Storage struct {
	mu   sync.Mutex
	pool *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.Connect(context.Background(), storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = pool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS url(
		id serial PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{mu: sync.Mutex{}, pool: pool}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"
	var id int64

	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id;`,
		alias,
		urlToSave,
	).Scan(&id)

	if err != nil {
		if pgerr, ok := err.(pgx.PgError); ok && pgerr.ConstraintName == "unique_alias" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	//TODO: refactor unique

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"
	var url string

	err := s.pool.QueryRow(context.Background(), `
		SELECT url
		FROM url
		WHERE alias = $1;
	`, alias).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"
	var id int

	err := s.pool.QueryRow(context.Background(), `
		DELETE FROM url
		WHERE alias = $1
		RETURNING id;
	`, alias).Scan(&id)

	if err != nil || id == 0 {
		if errors.Is(err, sql.ErrNoRows) || id == 0 {
			return storage.ErrURLNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
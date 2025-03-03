package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/stepan41k/FullRestAPI/internal/domain"
	"github.com/stepan41k/FullRestAPI/internal/storage"
)

type Storage struct {
	px *pgx.Conn
}

const (
	statusURLCreated = "URLCreated"
)

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	connect, err := pgx.Connect(context.Background(), storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = connect.Exec(context.Background(), `
	
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{px: connect}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (id int64, err error) {
	const op = "storage.postgres.SaveURL"

	tx, err := s.px.Begin(context.Background())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
			return
		}

		commitErr := tx.Commit(context.Background())
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, commitErr)
		}

	}()

	err = tx.QueryRow(context.Background(), `
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id;`,
		alias,
		urlToSave,
	).Scan(&id)

	if err != nil {
		// if pgerr, ok := err.(pgx.PgError); ok && pgerr.ConstraintName == "unique_alias" {
		// 	return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		// }

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	eventPayload := fmt.Sprintf(
		`{"id": %d, "url": "%s", "alias": "%s"}`,
		id,
		urlToSave,
		alias,
	)

	if err := s.saveEvent(tx, statusURLCreated, eventPayload); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) saveEvent(tx pgx.Tx, eventType string, payload string) error {
	const op = "storage.postgres.saveEvent"

	stmt, err := tx.Prepare(context.Background(), "my-query", "INSERT INTO events(event_type, payload) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(context.Background(), stmt.Name, eventType, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

type event struct {
	ID      int    `db:"id"`
	Type    string `db:"event_type"`
	Payload string `db:"payload"`
}

func (s *Storage) GetNewEvent(ctx context.Context) (domain.Event, error) {
	const op = "storage.postgres.GetNewEvent"

	//TODO: Научиться обрабатывать батчами
	row := s.px.QueryRow(context.Background(), `
		SELECT id, event_type, payload
		FROM events
		WHERE status = 'new'
		LIMIT 1
	`)

	var evt event

	err := row.Scan(&evt.ID, &evt.Type, &evt.Payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, nil
		}

		return domain.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Event{
		ID: evt.ID,
		Type: evt.Type,
		Payload: evt.Payload,
	}, nil

}

func (s *Storage) SetDone(id int) error  {
	const op = "storage.postgres.SetDone"

	stmt, err := s.px.Prepare(context.Background(), "set_done" ,`UPDATE events SET status = 'done' WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.px.Exec(context.Background(), stmt.Name, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"
	var url string

	err := s.px.QueryRow(context.Background(), `
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

	err := s.px.QueryRow(context.Background(), `
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

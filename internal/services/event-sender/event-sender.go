package eventsender

import (
	"context"
	"log/slog"
	"time"

	"github.com/stepan41k/FullRestAPI/internal/domain"
	"github.com/stepan41k/FullRestAPI/internal/lib/logger/sl"
	"github.com/stepan41k/FullRestAPI/internal/storage/postgres"
)

type Sender struct {
	storage *postgres.Storage
	log     *slog.Logger
}

func New(storage *postgres.Storage, log *slog.Logger) *Sender {
	return &Sender{
		storage: storage,
		log:     log,
	}
}

func (s *Sender) StartProcessEvents(ctx context.Context, handlePeriod time.Duration) {
	const op = "services.event-sender.StartProcessEvents"

	log := s.log.With(slog.String("op", op))

	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("stopping event processing")
				return
			case <-ticker.C:
				//noop
			}

			event, err := s.storage.GetNewEvent(ctx)
			if err != nil {
				if event.ID == 0 {
					log.Debug("no new events")
					continue
				}
				log.Error("failed to get new event", sl.Err(err))
			}

			s.SendMessage(event)

			if err := s.storage.SetDone(event.ID); err != nil {
				log.Error("failed to set event done", sl.Err(err))
				continue
			}
		}
	}()
}

func (s *Sender) SendMessage(event domain.Event) {
	const op = "services.event-sender.SendMessage"

	log := s.log.With(slog.String("op", op))
	log.Info("sending message", slog.Any("event", event))

	//TODO: implement sending message to the external sevice.
}

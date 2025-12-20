package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	"github.com/kostromin59/poster/internal/events"
	"github.com/kostromin59/poster/internal/models"
)

type PublishedPostTGPublisher interface {
	Publish(ctx context.Context, post events.PublishedPost) error
}

type PublishedPostTG struct {
	publisher PublishedPostTGPublisher
}

func NewPublishedPostTG(publisher PublishedPostTGPublisher) *PublishedPostTG {
	return &PublishedPostTG{
		publisher: publisher,
	}
}

func (ppt *PublishedPostTG) Handle(ctx context.Context, e []byte) {
	const op = "handlers.PublishedPostTG.Handle"

	log := slog.With(slog.String("op", op))

	var publishedPostEvent events.PublishedPost
	if err := json.Unmarshal(e, &publishedPostEvent); err != nil {
		log.Error("unable to unmarshal published post event", slog.String("err", err.Error()))
		return
	}

	if !slices.Contains(publishedPostEvent.Data.Sources, string(models.SourceTG)) {
		return
	}

	if err := ppt.publisher.Publish(ctx, publishedPostEvent); err != nil {
		log.Error("unable to publish tg post", slog.String("err", err.Error()))
		return
	}
}

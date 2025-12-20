package cronjob

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kostromin59/poster/internal/events"
	"github.com/kostromin59/poster/internal/models"
	"github.com/robfig/cron/v3"
)

type PublishedPostRepository interface {
	FindPublished(ctx context.Context, filters models.PostSearchFilters, offset, limit uint64) ([]models.Post, error)
}

func PublishedPost(ctx context.Context, cron *cron.Cron, d events.AsyncDispatcher, postRepo PublishedPostRepository) error {
	const op = "crojob.PublishedPost.PublishedPost"
	log := slog.With(slog.String("op", op))

	const limit = 10

	if _, err := cron.AddFunc("*/30 * * * *", func() {
		publishedFrom := time.Now().Add(-30 * time.Minute)
		var offset uint64

		for {
			posts, err := postRepo.FindPublished(ctx, models.PostSearchFilters{
				PublishedFrom: &publishedFrom,
			}, offset, limit)
			if err != nil {
				log.Error("unable to find published", slog.String("err", err.Error()))
				return
			}

			for _, p := range posts {
				eventID, err := uuid.NewRandom()
				if err != nil {
					log.Error("unable to generate event id", slog.String("err", err.Error()))
					continue
				}

				tags := make([]string, len(p.Tags))
				for i, t := range p.Tags {
					tags[i] = string(t)
				}

				sources := make([]string, len(p.Sources))
				for i, s := range p.Sources {
					sources[i] = string(s)
				}

				media := make([]events.PublishedPostMedia, len(p.Media))
				for i, m := range p.Media {
					media[i] = events.PublishedPostMedia{
						ID:       string(m.ID),
						Filetype: m.Filetype,
						URI:      m.URI,
					}
				}

				d.Dispatch(events.PublishedPost{
					EventID:   eventID.String(),
					CreatedAt: time.Now(),
					Data: events.PublishedPostData{
						ID:          string(p.ID),
						Title:       p.Title,
						Content:     p.Content,
						PublishDate: p.PublishDate,
						Tags:        tags,
						Media:       media,
						Sources:     sources,
					},
				})
			}

			if len(posts) < limit {
				break
			}

			offset += limit
		}
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

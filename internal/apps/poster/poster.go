package poster

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/configs"
	"github.com/kostromin59/poster/internal/events"
	"github.com/kostromin59/poster/internal/handlers"
	"github.com/kostromin59/poster/internal/infrastructure/cronjob"
	"github.com/kostromin59/poster/internal/infrastructure/dispatchers"
	"github.com/kostromin59/poster/internal/infrastructure/listeners"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/pkg/kafka"
	"github.com/robfig/cron/v3"
)

func Run(cfg configs.Poster) error {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Repositories
	pool, err := pgxpool.New(setupCtx, cfg.Database.DSN())
	if err != nil {
		return err
	}

	postRepo := pgxrepository.NewPost(pool)
	tagRepo := pgxrepository.NewTag(pool)
	_ = tagRepo

	// Kafka
	consumer, err := kafka.NewConsumer(cfg.KafkaHosts)
	if err != nil {
		return err
	}

	asyncProducer, err := kafka.NewAsyncProducer(cfg.KafkaHosts)
	if err != nil {
		return err
	}

	kafkaListener := listeners.NewKafka(consumer, cfg.PublishedPostTopic)
	kafkaListenerCh, err := kafkaListener.Start(appCtx)
	if err != nil {
		return err
	}

	// Handlers
	publishedPostTGHandler := handlers.NewPublishedPostTG(nil) // TODO:

	// Dispatchers
	publishedPostDispatcher := dispatchers.NewAsyncKakfa(asyncProducer, cfg.PublishedPostTopic)

	// Cron
	c := cron.New()
	defer c.Stop()
	if err := cronjob.PublishedPost(appCtx, c, publishedPostDispatcher, postRepo); err != nil {
		return err
	}
	c.Start()

	// Event listener
	go func() {
		handlers := []events.Handler{publishedPostTGHandler}
		for {
			select {
			case <-appCtx.Done():
				return
			case e, ok := <-kafkaListenerCh:
				if !ok {
					return
				}

				for _, h := range handlers {
					h.Handle(appCtx, e)
				}
			}
		}
	}()

	return nil
}

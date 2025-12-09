package poster

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kostromin59/poster/internal/configs"
	"github.com/kostromin59/poster/internal/events"
	"github.com/kostromin59/poster/internal/handlers"
	"github.com/kostromin59/poster/internal/infrastructure/cronjob"
	"github.com/kostromin59/poster/internal/infrastructure/dispatchers"
	"github.com/kostromin59/poster/internal/infrastructure/listeners"
	"github.com/kostromin59/poster/internal/infrastructure/pgxrepository"
	"github.com/kostromin59/poster/internal/infrastructure/tgbot"
	"github.com/kostromin59/poster/internal/models"
	"github.com/kostromin59/poster/pkg/kafka"
	"github.com/robfig/cron/v3"
	"gopkg.in/telebot.v4"
)

func Run(cfg *configs.Poster) error {
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

	// Event listeners
	kafkaPublishedPostListener := listeners.NewKafka(consumer, cfg.PublishedPostTopic)
	publisedPostCh, err := kafkaPublishedPostListener.Start(appCtx)
	if err != nil {
		return err
	}

	publishedPostListener := events.NewListener(publisedPostCh, publishedPostTGHandler)
	publishedPostListener.Start(appCtx)

	// Telegram bot
	telegramBot, err := telebot.NewBot(telebot.Settings{
		Token:     cfg.TGBotToken,
		ParseMode: telebot.ModeHTML,
		OnError: func(err error, c telebot.Context) {
			slog.Error("telegram bot error", slog.String("err", err.Error()))
			_ = c.Send("Что-то пошло не так! Попробуйте ещё раз!")
		},
	})
	if err != nil {
		return err
	}

	stepTG := tgbot.NewLocalState[string]()
	createPostState := tgbot.NewLocalState[models.CreatePostDTO]()
	createPostTGHandlers := tgbot.NewCreatePost(telegramBot, stepTG, createPostState, postRepo, tagRepo, nil)

	telegramBot.Use(tgbot.ContextMiddleware(), tgbot.CancelMiddleware(stepTG))
	telegramBot.Handle("/create_post", createPostTGHandlers.Handler())

	textHandlers := []telebot.HandlerFunc{
		createPostTGHandlers.TextAwaitingContentHandler(),
		createPostTGHandlers.TextAwaitingTitleHandler(),
	}

	telegramBot.Handle(telebot.OnText, func(c telebot.Context) error {
		for _, h := range textHandlers {
			if err := h(c); err != nil {
				return err
			}
		}

		return nil
	})

	slog.Info("app has been started")
	telegramBot.Start()

	return nil
}

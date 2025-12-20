package tgbot

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kostromin59/poster/internal/events"
	"gopkg.in/telebot.v4"
)

type PublisherCache interface {
	SetWithExpiration(ctx context.Context, key string, data []byte, exp time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

const PublisherAlreadyPublishedKey = "tgAlreadyPublished"

var PublisherCacheExpiration = 7 * 24 * time.Hour // Kafka data expiration

type Publisher struct {
	bot    *telebot.Bot
	chatID int64
	footer string
	cache  PublisherCache
}

func NewPublisher(bot *telebot.Bot, chatID int64, footer string, cache PublisherCache) *Publisher {
	return &Publisher{
		bot:    bot,
		chatID: chatID,
		footer: footer,
		cache:  cache,
	}
}

func (p *Publisher) Publish(ctx context.Context, post events.PublishedPost) error {
	const op = "tgbot.Publisher.Publish"

	alreadyPublishedRaw, err := p.cache.Get(ctx, PublisherAlreadyPublishedKey)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var alreadyPublished []string
	if len(alreadyPublishedRaw) != 0 {
		if err := json.Unmarshal(alreadyPublishedRaw, &alreadyPublished); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if slices.Contains(alreadyPublished, post.Data.ID) {
		return nil
	}

	msg := &strings.Builder{}
	msg.Grow(len(post.Data.Title) + len(post.Data.Title) + len(p.footer))

	if _, err := msg.WriteString("<b>"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString(post.Data.Title); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString("</b>\n\n"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString(post.Data.Content); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString("\n\n"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tags := strings.Join(post.Data.Tags, " ")

	if _, err := msg.WriteString(tags); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString("\n\n"); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := msg.WriteString(p.footer); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := p.bot.Send(&telebot.Chat{
		ID: p.chatID,
	}, msg.String(), telebot.ModeHTML); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	alreadyPublished = append(alreadyPublished, post.Data.ID)
	alreadyPublishedBytes, err := json.Marshal(alreadyPublished)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := p.cache.SetWithExpiration(ctx, PublisherAlreadyPublishedKey, alreadyPublishedBytes, PublisherCacheExpiration); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

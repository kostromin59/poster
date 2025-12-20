package tgbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/kostromin59/poster/internal/events"
	"gopkg.in/telebot.v4"
)

type Publisher struct {
	bot    *telebot.Bot
	chatID int64
	footer string
}

func NewPublisher(bot *telebot.Bot, chatID int64, footer string) *Publisher {
	return &Publisher{
		bot:    bot,
		chatID: chatID,
		footer: footer,
	}
}

func (p *Publisher) Publish(ctx context.Context, post events.PublishedPost) error {
	const op = "tgbot.Publisher.Publish"

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

	return nil
}

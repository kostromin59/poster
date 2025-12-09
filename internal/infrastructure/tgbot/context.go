package tgbot

import (
	"context"
	"time"

	"gopkg.in/telebot.v4"
)

const ContextKey = "ctx"

func ContextMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			c.Set(ContextKey, ctx)

			return next(c)
		}
	}
}

package tgbot

import (
	"slices"

	"gopkg.in/telebot.v4"
)

func AllowedUsersMiddleware(allowedIDs []int64) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if !slices.Contains(allowedIDs, c.Sender().ID) {
				return nil
			}

			return next(c)
		}
	}
}

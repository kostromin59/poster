package tgbot

import "gopkg.in/telebot.v4"

const CancelText = "Отменить"

func CancelMiddleware(step Step) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if c.Message().Text != CancelText {
				return next(c)
			}

			step.Delete(c.Sender().ID)

			kb := &telebot.ReplyMarkup{RemoveKeyboard: true}

			return c.Send("Действие отменено!", kb)
		}
	}
}

func CancelKeyboard() *telebot.ReplyMarkup {
	kb := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btn := kb.Text(CancelText)
	kb.Reply(kb.Row(btn))

	return kb
}

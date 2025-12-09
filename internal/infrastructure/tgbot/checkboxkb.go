package tgbot

import "gopkg.in/telebot.v4"

type CheckboxKeyboardItem struct {
	Value      string
	Label      string
	IsSelected bool
}

func NewCheckboxKeyboard(bot *telebot.Bot, prefix string, items []CheckboxKeyboardItem, edit func(c telebot.Context, items []CheckboxKeyboardItem) error) *telebot.ReplyMarkup {
	kb := bot.NewMarkup()

	buttons := make([]telebot.Row, 0, len(items))
	for i, item := range items {
		text := item.Label
		if item.IsSelected {
			text = "✅ " + text
		}

		btn := kb.Data(text, prefix, item.Value)

		bot.Handle(&btn, func(c telebot.Context) error {
			items[i].IsSelected = !items[i].IsSelected

			return edit(c, items)
		})

		buttons = append(buttons, kb.Row(btn))
	}

	kb.Inline(buttons...)

	return kb
}

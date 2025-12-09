package tgbot

import "gopkg.in/telebot.v4"

type CheckboxKeyboardItem struct {
	Value      string
	Label      string
	IsSelected bool
}

func CheckboxButtons(kb *telebot.ReplyMarkup, action string, items []CheckboxKeyboardItem) []telebot.Row {
	buttons := make([]telebot.Row, len(items))
	for i, item := range items {
		text := item.Label
		if item.IsSelected {
			text = "✅ " + text
		}

		btn := kb.Data(text, action, item.Value)
		buttons[i] = kb.Row(btn)
	}

	return buttons
}

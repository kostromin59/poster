package tgbot

import (
	"context"
	"errors"
	"strings"

	"github.com/kostromin59/poster/internal/models"
	"gopkg.in/telebot.v4"
)

type CreatePostRepository interface {
	Create(ctx context.Context, dto models.CreatePostDTO) (models.Post, error)
}

type CreatePostTagRepository interface {
	FindAll(ctx context.Context) ([]models.Tag, error)
}

type CreatePostSourceRepository interface {
	FindAll(ctx context.Context) ([]models.Source, error)
}

type CreatePostState struct {
	Title           string
	Content         string
	CheckboxTags    []CheckboxKeyboardItem
	Tags            []string
	CheckboxSources []CheckboxKeyboardItem
	Sources         []string
	Media           []string
}

type CreatePost struct {
	bot        *telebot.Bot
	step       Step
	state      State[CreatePostState]
	repo       CreatePostRepository
	tagRepo    CreatePostTagRepository
	sourceRepo CreatePostSourceRepository
}

func NewCreatePost(
	bot *telebot.Bot,
	step Step,
	state State[CreatePostState],
	repo CreatePostRepository,
	tagRepo CreatePostTagRepository,
	sourceRepo CreatePostSourceRepository,
) *CreatePost {
	return &CreatePost{
		bot:        bot,
		step:       step,
		state:      state,
		repo:       repo,
		tagRepo:    tagRepo,
		sourceRepo: sourceRepo,
	}
}

func (cp *CreatePost) Handler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		cp.step.Set(c.Sender().ID, StepAwaitingTitle)

		return c.Send("Введите заголовок:", CancelKeyboard())
	}
}

func (cp *CreatePost) TextAwaitingTitleHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingTitle {
			return nil
		}

		title := strings.TrimSpace(c.Message().Text)
		dto := CreatePostState{
			Title: title,
		}

		cp.state.Set(c.Sender().ID, dto)

		cp.step.Set(c.Sender().ID, StepAwaitingContent)

		return c.Send("Введите содержание:", CancelKeyboard())
	}
}

func (cp *CreatePost) TextAwaitingContentHandler() telebot.HandlerFunc {
	const actionToggleTag = "actionToggleTag"

	const message = "Добавьте уже существующие теги:"

	cp.bot.Handle("\f"+actionToggleTag, func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingTags {
			return nil
		}

		_ = c.Respond()

		value := c.Data()

		dto := cp.state.Get(c.Sender().ID)
		for i, ct := range dto.CheckboxTags {
			if ct.Value != value {
				continue
			}

			dto.CheckboxTags[i].IsSelected = !ct.IsSelected
		}

		cp.state.Set(c.Sender().ID, dto)

		kb := cp.bot.NewMarkup()
		checkboxButtons := CheckboxButtons(kb, actionToggleTag, dto.CheckboxTags)
		kb.Inline(checkboxButtons...)

		return c.Edit(message, kb)
	})

	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingContent {
			return nil
		}

		ctx := c.Get(ContextKey).(context.Context)

		content := strings.TrimSpace(c.Message().Text)
		dto := cp.state.Get(c.Sender().ID)
		dto.Content = content

		tags, err := cp.tagRepo.FindAll(ctx)
		if err != nil && !errors.Is(err, models.ErrTagNotFound) {
			return err
		}

		checkboxItems := make([]CheckboxKeyboardItem, len(tags))
		for i, tag := range tags {
			checkboxItems[i] = CheckboxKeyboardItem{
				Value:      string(tag),
				Label:      string(tag),
				IsSelected: false,
			}
		}
		dto.CheckboxTags = checkboxItems

		cp.state.Set(c.Sender().ID, dto)

		kb := cp.bot.NewMarkup()
		checkboxButtons := CheckboxButtons(kb, actionToggleTag, checkboxItems)
		kb.Inline(checkboxButtons...)

		cp.step.Set(c.Sender().ID, StepAwaitingTags)

		if err := c.Send("Напишите теги, если не хватает в списке.", CancelKeyboardWithButtons(NextStepButton)); err != nil {
			return err
		}

		return c.Send(message, kb)
	}
}

func (cp *CreatePost) TextAwaitingTagsHandler() telebot.HandlerFunc {
	const actionToggleSource = "actionToggleSource"

	const message = "Выберите источники:"

	cp.bot.Handle("\f"+actionToggleSource, func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingSources {
			return nil
		}

		_ = c.Respond()

		value := c.Data()

		dto := cp.state.Get(c.Sender().ID)
		for i, cs := range dto.CheckboxSources {
			if cs.Value != value {
				continue
			}

			dto.CheckboxSources[i].IsSelected = !cs.IsSelected
		}

		cp.state.Set(c.Sender().ID, dto)

		kb := cp.bot.NewMarkup()
		checkboxButtons := CheckboxButtons(kb, actionToggleSource, dto.CheckboxSources)
		kb.Inline(checkboxButtons...)

		return c.Edit(message, kb)
	})

	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingTags {
			return nil
		}

		ctx := c.Get(ContextKey).(context.Context)

		dto := cp.state.Get(c.Sender().ID)

		if c.Message().Text != NextStepButton {
			tagsRaw := strings.Split(c.Message().Text, ",")
			tags := make([]string, 0, len(tagsRaw))
			for _, tag := range tagsRaw {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}

				tags = append(tags, tag)
			}

			dto.Tags = tags
		}

		sources, err := cp.sourceRepo.FindAll(ctx)
		if err != nil && !errors.Is(err, models.ErrSourceNotFound) {
			return err
		}

		checkboxItems := make([]CheckboxKeyboardItem, len(sources))
		for i, s := range sources {
			checkboxItems[i] = CheckboxKeyboardItem{
				Value:      string(s),
				Label:      string(s),
				IsSelected: false,
			}
		}
		dto.CheckboxSources = checkboxItems

		kb := cp.bot.NewMarkup()
		checkboxButtons := CheckboxButtons(kb, actionToggleSource, checkboxItems)
		kb.Inline(checkboxButtons...)

		cp.state.Set(c.Sender().ID, dto)

		cp.step.Set(c.Sender().ID, StepAwaitingSources)

		if err := c.Send("Источники можно выбрать только из кнопок. После выбора нажмите кнопку «Продолжить»", CancelKeyboardWithButtons(NextStepButton)); err != nil {
			return err
		}

		return c.Send(message, kb)
	}
}

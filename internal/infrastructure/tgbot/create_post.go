package tgbot

import (
	"context"
	"errors"
	"strings"
	"time"

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
	PublishDate     time.Time
}

type CreatePost struct {
	bot        *telebot.Bot
	step       Step
	state      State[CreatePostState]
	repo       CreatePostRepository
	tagRepo    CreatePostTagRepository
	sourceRepo CreatePostSourceRepository
	loc        *time.Location
}

func NewCreatePost(
	bot *telebot.Bot,
	step Step,
	state State[CreatePostState],
	repo CreatePostRepository,
	tagRepo CreatePostTagRepository,
	sourceRepo CreatePostSourceRepository,
	loc *time.Location,
) *CreatePost {
	return &CreatePost{
		bot:        bot,
		step:       step,
		state:      state,
		repo:       repo,
		tagRepo:    tagRepo,
		sourceRepo: sourceRepo,
		loc:        loc,
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

func (cp *CreatePost) TextSubmitSourcesHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingSources {
			return nil
		}

		if c.Message().Text != NextStepButton {
			return nil
		}

		cp.step.Set(c.Sender().ID, StepAwaitingPublishDate)

		if err := c.Send("Введите дату публикации в формате 2006-01-02 15:04", CancelKeyboard()); err != nil {
			return err
		}

		return nil
	}
}

func (cp *CreatePost) TextAwaitingPublishDateHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingPublishDate {
			return nil
		}

		ctx := c.Get(ContextKey).(context.Context)

		publishDate, err := time.ParseInLocation("2006-01-02 15:04", c.Message().Text, cp.loc)
		if err != nil {
			if err := c.Reply("Не удалось разобрать дату публикации!"); err != nil {
				return err
			}

			return nil
		}

		dto := cp.state.Get(c.Sender().ID)
		dto.PublishDate = publishDate

		// TODO: media
		// cp.state.Set(c.Sender().ID, dto)

		tags := make([]models.Tag, 0, len(dto.Tags)+len(dto.CheckboxTags))
		for _, tag := range dto.Tags {
			if tag == "" {
				continue
			}

			tags = append(tags, models.Tag(tag))
		}

		for _, tag := range dto.CheckboxTags {
			if !tag.IsSelected {
				continue
			}

			if tag.Value == "" {
				continue
			}

			tags = append(tags, models.Tag(tag.Value))
		}

		sources := make([]models.Source, 0, len(dto.Sources)+len(dto.CheckboxSources))
		for i, s := range dto.Sources {
			if s == "" {
				continue
			}

			sources[i] = models.Source(s)
		}

		for _, s := range dto.CheckboxSources {
			if !s.IsSelected {
				continue
			}

			if s.Value == "" {
				continue
			}

			sources = append(sources, models.Source(s.Value))
		}

		if _, err := cp.repo.Create(ctx, models.CreatePostDTO{
			Title:       dto.Title,
			Content:     dto.Content,
			PublishDate: dto.PublishDate,
			Tags:        tags,
			Sources:     sources,
			Media:       nil,
		}); err != nil {
			return err
		}

		kb := &telebot.ReplyMarkup{RemoveKeyboard: true}

		cp.step.Delete(c.Sender().ID)
		cp.state.Delete(c.Sender().ID)

		if err := c.Send("Пост создан!", kb); err != nil {
			return err
		}

		return nil
	}
}

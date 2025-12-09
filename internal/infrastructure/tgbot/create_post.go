package tgbot

import (
	"context"
	"maps"
	"slices"
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

type CreatePost struct {
	bot        *telebot.Bot
	step       Step
	state      State[models.CreatePostDTO]
	repo       CreatePostRepository
	tagRepo    CreatePostTagRepository
	sourceRepo CreatePostSourceRepository
}

func NewCreatePost(
	bot *telebot.Bot,
	step Step,
	state State[models.CreatePostDTO],
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
		dto := models.CreatePostDTO{
			Title: title,
		}

		cp.state.Set(c.Sender().ID, dto)

		cp.step.Set(c.Sender().ID, StepAwaitingContent)

		return c.Send("Введите содержание:", CancelKeyboard())
	}
}

func (cp *CreatePost) TextAwaitingContentHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingContent {
			return nil
		}

		ctx := c.Get(ContextKey).(context.Context)

		content := strings.TrimSpace(c.Message().Text)
		dto := cp.state.Get(c.Sender().ID)
		dto.Content = content

		cp.state.Set(c.Sender().ID, dto)

		tags, err := cp.tagRepo.FindAll(ctx)
		if err != nil {
			return err
		}

		radioItems := make([]CheckboxKeyboardItem, len(tags))
		for i, tag := range tags {
			radioItems[i] = CheckboxKeyboardItem{
				Value:      string(tag),
				Label:      string(tag),
				IsSelected: false,
			}
		}

		var edit func(c telebot.Context, items []CheckboxKeyboardItem) error
		edit = func(c telebot.Context, items []CheckboxKeyboardItem) error {
			tags := make([]string, 0, len(items))
			for _, item := range items {
				if item.IsSelected {
					tags = append(tags, item.Value)
				}
			}

			cp.updateStateTags(c.Sender().ID, tags)

			return c.Edit("Введите теги:", CancelKeyboard(), NewCheckboxKeyboard(cp.bot, "tags", radioItems, edit))
		}

		cp.step.Set(c.Sender().ID, StepAwaitingTags)

		return c.Send("Введите теги:", CancelKeyboard(), NewCheckboxKeyboard(cp.bot, "tags", radioItems, edit))
	}
}

func (cp *CreatePost) updateStateTags(userID int64, tags []string) {
	dto := cp.state.Get(userID)

	uniqueTags := make(map[models.Tag]struct{}, len(dto.Tags)+len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		uniqueTags[models.Tag(tag)] = struct{}{}
	}

	for _, tag := range dto.Tags {
		uniqueTags[tag] = struct{}{}
	}

	allTags := slices.Collect(maps.Keys(uniqueTags))
	dto.Tags = allTags

	cp.state.Set(userID, dto)
}

func (cp *CreatePost) TextAwaitingTagsHandler() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if cp.step.Get(c.Sender().ID) != StepAwaitingTags {
			return nil
		}

		title := strings.TrimSpace(c.Message().Text)
		dto := models.CreatePostDTO{
			Title: title,
		}

		cp.state.Set(c.Sender().ID, dto)

		// cp.step.Set(c.Sender().ID, StepAwaitingTags)

		return c.Send("Введите источники:", CancelKeyboard())
	}
}

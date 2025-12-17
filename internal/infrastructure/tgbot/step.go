package tgbot

const (
	StepAwaitingTitle       = "awaitingTitle"
	StepAwaitingContent     = "awaitingContent"
	StepAwaitingTags        = "awaitingTags"
	StepAwaitingSources     = "awaitingSources"
	StepAwaitingPublishDate = "awaitingPublishDate"
)

const NextStepButton = "Продолжить"

type Step interface {
	Get(userID int64) string
	Set(userID int64, step string)
	Delete(userID int64)
}

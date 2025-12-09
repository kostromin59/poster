package tgbot

const (
	StepAwaitingTitle   = "awaitingTitle"
	StepAwaitingContent = "awaitingContent"
	StepAwaitingTags    = "awaitingTags"
	StepAwaitingSources = "awaitingSources"
)

const NextStepButton = "Продолжить"

type Step interface {
	Get(userID int64) string
	Set(userID int64, step string)
	Delete(userID int64)
}

// type LocalStep struct {
// 	m  map[int64]string
// 	mu sync.RWMutex
// }

// func NewLocalStep() *LocalStep {
// 	return &LocalStep{
// 		m: make(map[int64]string),
// 	}
// }

// func (ls *LocalStep) Get(userID int64) string {
// 	ls.mu.RLock()
// 	step := ls.m[userID]
// 	ls.mu.RUnlock()

// 	return step
// }

// func (ls *LocalStep) Set(userID int64, step string) {
// 	ls.mu.Lock()
// 	ls.m[userID] = step
// 	ls.mu.Unlock()
// }

// func (ls *LocalStep) Delete(userID int64) {
// 	ls.mu.Lock()
// 	delete(ls.m, userID)
// 	ls.mu.Unlock()
// }

package tgbot

import "sync"

type State[T any] interface {
	Set(userID int64, data T)
	Get(userID int64) T
	Delete(userID int64)
}

type LocalState[T any] struct {
	m  map[int64]T
	mu sync.RWMutex
}

func NewLocalState[T any]() *LocalState[T] {
	return &LocalState[T]{
		m: make(map[int64]T),
	}
}

func (ls *LocalState[T]) Get(userID int64) T {
	ls.mu.RLock()
	step := ls.m[userID]
	ls.mu.RUnlock()

	return step
}

func (ls *LocalState[T]) Set(userID int64, data T) {
	ls.mu.Lock()
	ls.m[userID] = data
	ls.mu.Unlock()
}

func (ls *LocalState[T]) Delete(userID int64) {
	ls.mu.Lock()
	delete(ls.m, userID)
	ls.mu.Unlock()
}

package events

import "context"

type Listener struct {
	handlers []Handler
	ch       <-chan []byte
}

func NewListener(ch <-chan []byte, handlers ...Handler) *Listener {
	return &Listener{
		handlers: handlers,
		ch:       ch,
	}
}

func (l *Listener) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-l.ch:
				if !ok {
					return
				}

				for _, h := range l.handlers {
					h.Handle(ctx, event)
				}
			}
		}
	}()
}

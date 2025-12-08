package events

import "context"

type Handler interface {
	Handle(context.Context, []byte)
}

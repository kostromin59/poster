package events

type AsyncDispatcher interface {
	Dispatch(any)
}

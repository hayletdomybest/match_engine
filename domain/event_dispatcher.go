package domain

type EventDispatcher interface {
	Dispatch(event MatchEvent)
}

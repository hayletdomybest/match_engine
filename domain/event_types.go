package domain

type MatchEventType int

const (
	OrderCreated MatchEventType = iota
)

type MatchEvent struct {
	Type MatchEventType
	Data any
}

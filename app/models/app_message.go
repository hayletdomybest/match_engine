package models

type AppMessage[T any] struct {
	Action string
	Data   T
}

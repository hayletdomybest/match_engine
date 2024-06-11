package domain

import "time"

type OrderCreatedEvent struct {
	OrderID   string
	CreatedAt int64
}

func NewOrderCreatedEvent(orderID string) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		OrderID:   orderID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

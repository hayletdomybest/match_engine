package order

import (
	"errors"
	"time"
)

type Order struct {
	Id             uint64    `json:"id"`
	Symbol         string    `json:"symbol"`
	Type           OrderType `json:"type"`
	Quantity       float64   `json:"quantity"`
	FilledQuantity float64   `json:"filled_quantity"`
	Price          float64   `json:"price"`
	Side           Side      `json:"side"`
	Status         Status    `json:"status"`
	Timestamp      int64     `json:"timestamp"`
}

func NewOrder(
	id uint64, symbol string,
	orderType OrderType, price float64, quantity float64, side Side) (*Order, error) {
	if len(symbol) == 0 || quantity == 0 || price == 0 {
		return nil, errors.New("missing required fields")
	}
	return &Order{
		Id:        id,
		Symbol:    symbol,
		Type:      orderType,
		Quantity:  quantity,
		Price:     price,
		Side:      side,
		Status:    Pending,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

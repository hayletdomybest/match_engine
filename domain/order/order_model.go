package order

import (
	"errors"
	"time"
)

type Order struct {
	Id             uint64    `json:"id"`
	MarketId       uint64    `json:"market_id"`
	MarketCode     string    `json:"market_code"`
	Type           OrderType `json:"type"`
	Quantity       float64   `json:"quantity"`
	FilledQuantity float64   `json:"filled_quantity"`
	Price          float64   `json:"price"`
	Direction      Direction `json:"direction"`
	Status         Status    `json:"status"`
	Timestamp      int64     `json:"timestamp"`
}

func NewOrder(
	id uint64, marketId uint64, marketCode string,
	orderType OrderType, quantity float64, price float64, direction Direction) (*Order, error) {
	if marketId == 0 || marketCode == "" || quantity == 0 || price == 0 {
		return nil, errors.New("missing required fields")
	}
	return &Order{
		Id:         id,
		MarketId:   marketId,
		MarketCode: marketCode,
		Type:       orderType,
		Quantity:   quantity,
		Price:      price,
		Direction:  direction,
		Status:     New,
		Timestamp:  time.Now().UnixMilli(),
	}, nil
}

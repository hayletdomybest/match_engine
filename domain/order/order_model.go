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
	Side           Side      `json:"direction"`
	Status         Status    `json:"status"`
	Timestamp      int64     `json:"timestamp"`
}

func NewOrder(
	id uint64, marketId uint64, marketCode string,
	orderType OrderType, price float64, quantity float64, direction Side) (*Order, error) {
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
		Side:       direction,
		Status:     Pending,
		Timestamp:  time.Now().UnixMilli(),
	}, nil
}

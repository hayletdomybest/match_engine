package db

import (
	"encoding/json"
	"match_engine/domain"
	odr "match_engine/domain/order"
)

type OrderBookKv struct {
	state map[string]any
}

func NewOrderBookKV() *OrderBookKv {
	return &OrderBookKv{
		state: make(map[string]any),
	}
}

func (book *OrderBookKv) GetOrderBookValue(code string, direction domain.Direction) []odr.Order {
	key := OrderBookValueKey(code, direction)

	val, existed := book.state[key]
	if !existed {
		return []odr.Order{}
	}

	return val.([]odr.Order)
}

func (book *OrderBookKv) GetOrderBookCount(code string, direction domain.Direction) uint64 {
	key := OrderBookCountKey(code, direction)
	val, existed := book.state[key]
	if !existed {
		return 0
	}

	return val.(uint64)
}

func (book *OrderBookKv) AddOrder(order odr.Order) {
	key := OrderBookValueKey(order.MarketCode, order.Direction)

	_, existed := book.state[key]
	if !existed {
		var orders []odr.Order
		book.state[key] = orders
	}

	book.state[key] = append(book.state[key].([]odr.Order), order)
}

func (book *OrderBookKv) AddOrders(orders ...odr.Order) {
	for _, order := range orders {
		book.AddOrder(order)
	}
}

func (book *OrderBookKv) SetOrderCount(code string, direction domain.Direction, count uint64) {
	key := OrderBookCountKey(code, direction)
	book.state[key] = count
}

func (book *OrderBookKv) LoadSnap(bz []byte) {
	if len(bz) == 0 {
		return
	}

	var restore map[string]any
	json.Unmarshal(bz, &restore)

	book.state = restore
}

func (book *OrderBookKv) CreateSnap() []byte {
	if len(book.state) == 0 {
		return nil
	}
	bz, _ := json.Marshal(book.state)
	return bz
}

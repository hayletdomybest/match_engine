package db

import (
	"encoding/json"
	"match_engine/domain"
	odr "match_engine/domain/order"
)

type OrderSnapshot struct {
	Value map[string][]odr.Order
	Count map[string]uint64
}

type OrderBookKv struct {
	value map[string][]odr.Order
	count map[string]uint64
}

func NewOrderBookKV() *OrderBookKv {
	return &OrderBookKv{
		value: make(map[string][]odr.Order),
		count: make(map[string]uint64),
	}
}

func (book *OrderBookKv) GetOrderBookValue(code string, direction domain.Direction) []odr.Order {
	key := OrderBookValueKey(code, direction)

	val, existed := book.value[key]
	if !existed {
		return []odr.Order{}
	}

	return val
}

func (book *OrderBookKv) GetOrderBookCount(code string, direction domain.Direction) uint64 {
	key := OrderBookCountKey(code, direction)
	val, existed := book.count[key]
	if !existed {
		return 0
	}

	return val
}

func (book *OrderBookKv) AddOrder(order odr.Order) {
	key := OrderBookValueKey(order.MarketCode, order.Direction)

	_, existed := book.value[key]
	if !existed {
		var orders []odr.Order
		book.value[key] = orders
	}

	book.value[key] = append(book.value[key], order)
}

func (book *OrderBookKv) AddOrders(orders ...odr.Order) {
	for _, order := range orders {
		book.AddOrder(order)
	}
}

func (book *OrderBookKv) SetOrderCount(code string, direction domain.Direction, count uint64) {
	key := OrderBookCountKey(code, direction)
	book.count[key] = count
}

func (book *OrderBookKv) LoadSnap(bz []byte) {
	if len(bz) == 0 {
		return
	}

	var restore OrderSnapshot
	json.Unmarshal(bz, &restore)

	for k, v := range restore.Value {
		book.value[k] = v
	}

	for k, v := range restore.Count {
		book.count[k] = v
	}
}

func (book *OrderBookKv) CreateSnap() []byte {
	if len(book.value) == 0 && len(book.count) == 0 {
		return nil
	}

	snap := OrderSnapshot{}
	snap.Value = book.value
	snap.Count = book.count

	bz, _ := json.Marshal(snap)
	return bz
}

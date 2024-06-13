package db

import (
	"encoding/json"
	odr "match_engine/domain/order"
	"match_engine/utils"
	"sort"
)

type MatchEngineKv struct {
	Order     map[uint64]*odr.Order
	Orderbook map[string][]*odr.Order
}

func NewMatchEngineKV() *MatchEngineKv {
	return &MatchEngineKv{
		Order:     make(map[uint64]*odr.Order),
		Orderbook: make(map[string][]*odr.Order),
	}
}

func (kv *MatchEngineKv) Save(orders ...odr.Order) error {
	for _, order := range orders {
		kv.save(&order)
	}
	return nil
}

func (kv *MatchEngineKv) FindById(orderId uint64) (*odr.Order, error) {
	order, existed := kv.Order[orderId]
	if !existed {
		return nil, nil
	}

	return order, nil
}

func (kv *MatchEngineKv) FetchMatchingOrders(order odr.Order, total uint64) ([]odr.Order, error) {
	key := OrderBookValueKey(order.MarketCode, odr.OppositeSide(order.Side))
	orders, existed := kv.Orderbook[key]
	if !existed {
		return []odr.Order{}, nil
	}

	var count uint64 = 0
	for _, matchedOrder := range orders {
		if !odr.NewOrderAggregate(&order).CanMatch(matchedOrder) {
			break
		}
		count++
		if total != 0 && count == total {
			break
		}
	}

	rest, matched := utils.Shift(kv.Orderbook[key], int(count))
	kv.Orderbook[key] = rest

	return utils.Select(matched, func(item *odr.Order) odr.Order {
		return *item
	}), nil
}

func (kv *MatchEngineKv) LoadSnap(bz []byte) {
	if len(bz) == 0 {
		return
	}

	var restore MatchEngineKv
	json.Unmarshal(bz, &restore)

	kv.Order = restore.Order
	kv.Orderbook = restore.Orderbook
}

func (book *MatchEngineKv) CreateSnap() []byte {
	if len(book.Orderbook) == 0 && len(book.Order) == 0 {
		return nil
	}

	bz, _ := json.Marshal(book)
	return bz
}

func (kv *MatchEngineKv) appendOrderbook(order *odr.Order) error {
	key := OrderBookValueKey(order.MarketCode, order.Side)
	direction := order.Side

	_, existed := kv.Orderbook[key]
	if !existed {
		var orders []*odr.Order
		kv.Orderbook[key] = orders
	}

	kv.Orderbook[key] = append(kv.Orderbook[key], order)

	sort.Slice(kv.Orderbook[key], func(i, j int) bool {
		iOrder := kv.Orderbook[key][i]
		jOrder := kv.Orderbook[key][j]

		if direction == odr.Sell {
			return iOrder.Price <= jOrder.Price
		} else {
			return iOrder.Price >= jOrder.Price
		}
	})

	return nil
}

func (kv *MatchEngineKv) save(order *odr.Order) {
	kv.Order[order.Id] = order

	if !order.IsDone() {
		kv.appendOrderbook(order)
	}
}

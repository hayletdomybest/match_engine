package db_test

import (
	"encoding/json"
	"match_engine/domain/market"
	odr "match_engine/domain/order"
	"match_engine/infra/db"
	"match_engine/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

var DefaultMarkets = []market.Market{
	{
		Symbol: "BTCUSDT",
	},
	{
		Symbol: "ETHUSDT",
	},
}

func TestSave(t *testing.T) {
	kv := db.NewMatchEngineKV()

	market := DefaultMarkets[0]

	order, _ := odr.NewOrder(1, market.Symbol, odr.Limit, 10, 100, odr.Buy)
	kv.Save(*order)

	find, _ := kv.FindById(order.Id)

	bz1, _ := json.Marshal(find)
	bz2, _ := json.Marshal(*order)
	assert.Equal(t, string(bz2), string(bz1))
}

func TestFetchMatchingOrders(t *testing.T) {
	kv := db.NewMatchEngineKV()

	market := DefaultMarkets[0]

	order, _ := odr.NewOrder(1, market.Symbol, odr.Limit, 100, 10, odr.Buy)
	kv.Save(*order)
	order, _ = odr.NewOrder(2, market.Symbol, odr.Limit, 101, 10, odr.Buy)
	kv.Save(*order)
	order, _ = odr.NewOrder(3, market.Symbol, odr.Limit, 102, 10, odr.Buy)
	kv.Save(*order)

	sell, _ := odr.NewOrder(4, market.Symbol, odr.Limit, 101, 20, odr.Sell)
	find, _ := kv.FetchMatchingOrders(*sell, 3)

	assert.Equal(t, 2, len(find))

	assert.Contains(t, []int{101, 102}, utils.Select(find, func(o odr.Order) int { return int(o.Id) }))
}

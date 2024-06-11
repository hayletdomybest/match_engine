package db_test

import (
	"encoding/json"
	"match_engine/app"
	"match_engine/domain"
	odr "match_engine/domain/order"
	"match_engine/infra/db"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddOrder(t *testing.T) {
	orderbook := db.NewOrderBookKV()

	market := app.DefaultMarkets[0]

	order, _ := odr.NewOrder(1, market.Id, market.Code, domain.Limit, 10, 100, domain.Bid)
	orderbook.AddOrders(*order)

	getOrders := orderbook.GetOrderBookValue(market.Code, domain.Bid)

	assert.Equal(t, 1, len(getOrders))

	bz1, _ := json.Marshal(getOrders[0])
	bz2, _ := json.Marshal(*order)
	assert.Equal(t, string(bz2), string(bz1))
}

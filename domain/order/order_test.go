package order_test

import (
	"match_engine/domain"
	odr "match_engine/domain/order"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilledOrder(t *testing.T) {
	order := odr.NewOrderAggregate(odr.Order{
		Id:             1,
		MarketId:       1,
		MarketCode:     "BTC-USD",
		Type:           domain.Limit,
		Quantity:       10.5,
		FilledQuantity: 10.5,
		Price:          45000.0,
		Direction:      domain.Ask,
		Status:         odr.New,
	})
	order.Done()

	assert.Equal(t, order.Status, odr.FullFiled)
}

func TestShouldPanic(t *testing.T) {
	order := odr.NewOrderAggregate(odr.Order{
		Id:             1,
		MarketId:       1,
		MarketCode:     "BTC-USD",
		Type:           domain.Limit,
		Quantity:       10.5,
		FilledQuantity: 0,
		Price:          45000.0,
		Direction:      domain.Ask,
		Status:         odr.New,
	})
	assert.Panics(t, func() {
		order.Done()
	})
}

func TestMarshal(t *testing.T) {
	order := odr.NewOrderAggregate(odr.Order{
		Id:             1,
		MarketId:       1,
		MarketCode:     "BTC-USD",
		Type:           domain.Limit,
		Quantity:       10.5,
		FilledQuantity: 10.5,
		Price:          45000.0,
		Direction:      domain.Ask,
		Status:         odr.New,
	})

	bz := order.Marshal()
	str := string(bz)
	assert.Equal(t, "{\"id\":1,\"market_id\":1,\"market_code\":\"BTC-USD\",\"type\":0,\"amount\":10.5,\"filled_amount\":10.5,\"price\":45000,\"direction\":1,\"status\":0}", str)
}

package order_test

import (
	odr "match_engine/domain/order"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilledOrder(t *testing.T) {
	order := odr.NewOrderAggregate(&odr.Order{
		Id:             1,
		Symbol:         "BTC-USD",
		Type:           odr.Limit,
		Quantity:       10.5,
		FilledQuantity: 10.5,
		Price:          45000.0,
		Side:           odr.Sell,
		Status:         odr.Pending,
	})
	order.Done()

	assert.Equal(t, order.Status, odr.FullFiled)
}

func TestShouldPanic(t *testing.T) {
	order := odr.NewOrderAggregate(&odr.Order{
		Id:             1,
		Symbol:         "BTC-USD",
		Type:           odr.Limit,
		Quantity:       10.5,
		FilledQuantity: 0,
		Price:          45000.0,
		Side:           odr.Sell,
		Status:         odr.Pending,
	})
	assert.Panics(t, func() {
		order.Done()
	})
}

func TestMarshal(t *testing.T) {
	order := odr.NewOrderAggregate(&odr.Order{
		Id:             1,
		Symbol:         "BTC-USD",
		Type:           odr.Limit,
		Quantity:       10.5,
		FilledQuantity: 10.5,
		Price:          45000.0,
		Side:           odr.Sell,
		Status:         odr.Pending,
	})

	bz := order.Marshal()
	str := string(bz)
	assert.Equal(t, "{\"id\":1,\"symbol\":\"BTC-USD\",\"type\":0,\"quantity\":10.5,\"filled_quantity\":10.5,\"price\":45000,\"side\":1,\"status\":0,\"timestamp\":0}", str)
}

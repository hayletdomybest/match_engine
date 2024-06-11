package order

import (
	"encoding/json"
	"sync"
)

type OrderAggregateRoot struct {
	Order
	mu sync.Mutex `json:"-"`
}

func NewOrderAggregate(order Order) *OrderAggregateRoot {
	return &OrderAggregateRoot{
		Order: order,
	}
}

func (order *OrderAggregateRoot) Done() {
	order.done(false)
}

func (order *OrderAggregateRoot) IsDone() bool {
	return order.Status.IsDone()
}

func (order *OrderAggregateRoot) Cancel() {
	order.done(true)
}

func (order *OrderAggregateRoot) Marshal() []byte {
	bz, _ := json.Marshal(order.Order)
	return bz
}

func (order *OrderAggregateRoot) done(canceled bool) {
	if order.IsDone() {
		return
	}

	defer order.mu.Unlock()
	order.mu.Lock()

	if order.IsDone() {
		return
	}
	if order.FilledQuantity == 0 && !canceled {
		panic("can not finish order on zero filled")
	}

	if order.FilledQuantity == 0 {
		order.Status = Canceled
		return
	}

	left := order.Quantity - order.FilledQuantity
	if left == 0 {
		order.Status = FullFiled
	} else {
		order.Status = PartialFilled
	}
}

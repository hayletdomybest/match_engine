package order

import (
	"encoding/json"
)

type odrOrderAggregate struct {
	*Order
}

func NewOrderAggregate(order *Order) *odrOrderAggregate {
	return &odrOrderAggregate{
		Order: order,
	}
}

func (aggregate *odrOrderAggregate) Done() {
	aggregate.done(false)
}

func (aggregate *odrOrderAggregate) IsDone() bool {
	return aggregate.Order.IsDone()
}

func (aggregate *odrOrderAggregate) Cancel() {
	aggregate.done(true)
}

func (aggregate *odrOrderAggregate) Marshal() []byte {
	bz, _ := json.Marshal(aggregate.Order)
	return bz
}

func (aggregate *odrOrderAggregate) CanMatch(matchedOrder *Order) bool {
	if aggregate.IsDone() || matchedOrder.IsDone() || aggregate.Side == matchedOrder.Side {
		return false
	}

	if aggregate.Side == Sell {
		return matchedOrder.Price >= aggregate.Price
	} else {
		return matchedOrder.Price <= aggregate.Price
	}
}

func (aggregate *odrOrderAggregate) Match(matchedOrders ...*Order) []*Order {
	var res []*Order

	for _, matchedOrder := range matchedOrders {
		if aggregate.IsDone() {
			break
		}

		if aggregate.Fill(matchedOrder) {
			res = append(res, matchedOrder)
			continue
		}
	}

	return res
}

func (aggregate *odrOrderAggregate) Fill(matchedOrder *Order) bool {
	if !aggregate.CanMatch(matchedOrder) {
		return false
	}

	var filledAmount float64
	if aggregate.UnFilled() >= NewOrderAggregate(matchedOrder).UnFilled() {
		filledAmount = NewOrderAggregate(matchedOrder).UnFilled()
	} else {
		filledAmount = aggregate.UnFilled()
	}
	aggregate.FilledQuantity += filledAmount
	NewOrderAggregate(matchedOrder).FilledQuantity += filledAmount
	return true
}

func (aggregate *odrOrderAggregate) UnFilled() float64 {
	if aggregate.IsDone() {
		return 0
	}

	return aggregate.Quantity - aggregate.FilledQuantity
}

func (aggregate *odrOrderAggregate) done(canceled bool) {
	if aggregate.IsDone() {
		return
	}

	if aggregate.IsDone() {
		return
	}
	if aggregate.FilledQuantity == 0 && !canceled {
		panic("can not finish order on zero filled")
	}

	if aggregate.FilledQuantity == 0 {
		aggregate.Status = Canceled
		return
	}

	left := aggregate.Quantity - aggregate.FilledQuantity
	if left == 0 {
		aggregate.Status = FullFiled
	} else {
		aggregate.Status = PartialFilled
	}
}

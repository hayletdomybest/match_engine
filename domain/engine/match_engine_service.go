package engine

import (
	"match_engine/domain"
	mk "match_engine/domain/market"
	odr "match_engine/domain/order"
	"match_engine/utils"
)

type MatchEngine struct {
	orderRepo  odr.OrderRepository
	marketRepo mk.MarketRepository

	eventDispatcher domain.EventDispatcher
}

func NewMatchEngin(
	orderRepo odr.OrderRepository,
	marketRepo mk.MarketRepository,
	eventDispatcher domain.EventDispatcher) *MatchEngine {
	return &MatchEngine{
		orderRepo:       orderRepo,
		marketRepo:      marketRepo,
		eventDispatcher: eventDispatcher,
	}
}

func (me *MatchEngine) Match(order *odr.Order) MatchResult {
	//TODO
	var matched []odr.Order
	var result MatchResult
	for !order.IsDone() {
		matchedOrder, _ := me.orderRepo.FetchMatchingOrders(*order, 1)

		if len(matchedOrder) > 0 {
			matched = append(matched, matchedOrder...)
		}

		aggregate := odr.NewOrderAggregate(order)
		aggregate.Match(utils.Select(matchedOrder, func(o odr.Order) *odr.Order { return &o }))

		for _, mo := range matchedOrder {
			order.FilledQuantity += mo.Quantity
		}

		if order.FilledQuantity >= order.Quantity {
			order.Status = odr.FullFiled
		}
	}

	me.orderRepo.Save(append(matched, *order)...)

	//TODO
	//me.eventDispatcher.Dispatch()

	result = MatchResult{
		Order:   order,
		Matched: matched,
	}

	return result
}

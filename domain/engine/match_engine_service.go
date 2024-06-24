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

func (me *MatchEngine) Match(order *odr.Order) *MatchResult {
	result := &MatchResult{
		Order: order,
	}
	aggregate := odr.NewOrderAggregate(order)
	for !aggregate.IsDone() {
		matchedOrder, _ := me.orderRepo.FetchMatchingOrders(*order, 1)

		if len(matchedOrder) > 0 {
			break
		}

		matched := aggregate.Match(&matchedOrder[0])
		result.Matched = append(result.Matched, matched...)
	}

	me.orderRepo.Save(
		utils.Select(append(result.Matched, order), func(o *odr.Order) odr.Order { return *o })...,
	)

	//TODO
	//me.eventDispatcher.Dispatch()

	return result
}

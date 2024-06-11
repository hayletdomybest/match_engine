package engine

import (
	mk "match_engine/domain/market"
	odr "match_engine/domain/order"
)

type MatchEngine struct {
	orderRepo  odr.OrderRepository
	marketRepo mk.MarketRepository
}

func NewMatchEngin(orderRepo odr.OrderRepository, marketRepo mk.MarketRepository) *MatchEngine {
	return &MatchEngine{
		orderRepo:  orderRepo,
		marketRepo: marketRepo,
	}
}

func Match(order *odr.Order) {
	//TODO
}

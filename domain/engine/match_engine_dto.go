package engine

import (
	odr "match_engine/domain/order"
)

type MatchResult struct {
	Order   *odr.Order
	Matched []*odr.Order
}

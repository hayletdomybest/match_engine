package db

import (
	"fmt"
	odr "match_engine/domain/order"
)

func OrderBookValueKey(code string, dir odr.Side) string {
	return fmt.Sprintf("/orderbook/%s/%s/value", code, dir.ToString())
}

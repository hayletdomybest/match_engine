package db

import (
	"fmt"
	"match_engine/domain"
)

func OrderBookCountKey(code string, dir domain.Direction) string {
	return fmt.Sprintf("/orderbook/%s/%s/count", code, dir.ToString())
}

func OrderBookValueKey(code string, dir domain.Direction) string {
	return fmt.Sprintf("/orderbook/%s/%s/value", code, dir.ToString())
}

package app

import "match_engine/domain/market"

var DefaultMarkets = []market.Market{
	{
		Id:   1,
		Code: "BTCUSDT",
	},
	{
		Id:   2,
		Code: "ETHUSDT",
	},
}

package app

import "match_engine/domain/market"

var DefaultMarkets = []market.Market{
	{
		Symbol: "BTCUSDT",
	},
	{
		Symbol: "ETHUSDT",
	},
}

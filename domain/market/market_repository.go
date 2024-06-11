package market

type MarketRepository interface {
	Save(market *Market) error
	FindById(marketId string) (*Market, error)
	FindByCode(marketId string) (*Market, error)
}

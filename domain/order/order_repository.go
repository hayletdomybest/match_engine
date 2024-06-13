package order

type OrderRepository interface {
	Save(...Order) error
	FindById(uint64) (*Order, error)
	FetchMatchingOrders(Order, uint64) ([]Order, error)
}

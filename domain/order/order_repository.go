package order

type OrderRepository interface {
	Save(order *Order) error
	FindById(orderId string) (*Order, error)
	FindMatchingOrders(order *Order) ([]*Order, error)
}

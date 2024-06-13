package order

func (order *Order) IsDone() bool {
	return order.Status.IsDone()
}

func OppositeSide(side Side) Side {
	if side == Buy {
		return Sell
	} else {
		return Buy
	}
}

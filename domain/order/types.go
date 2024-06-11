package order

type Status int

const (
	New Status = iota
	Canceled
	FullFiled
	PartialFilled
)

var strStatus = []string{"new", "canceled", "fullFilled", "partialFilled"}

func (s Status) IsDone() bool {
	return s != New
}

func (s Status) ToString() string {
	if int(s) < 0 || int(s) >= len(strStatus) {
		panic("order status out of range")
	}

	return strStatus[s]
}

type Direction int

const (
	Bid Direction = iota
	Ask
)

var strDirection = []string{"bid", "ask"}

func (d Direction) ToString() string {
	if int(d) >= len(strDirection) || int(d) < 0 {
		panic("direction out of range")
	}
	return strDirection[d]
}

type OrderType int

const (
	Limit OrderType = iota
	Market
)

var strOrderType = []string{"limit", "market"}

func (d OrderType) ToString() string {
	if int(d) >= len(strOrderType) || int(d) < 0 {
		panic("order type out of range")
	}
	return strOrderType[d]
}

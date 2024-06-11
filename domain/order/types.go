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

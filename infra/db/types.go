package db

type OrderBookKvSnapshot map[string]any

type Repository interface {
	GetName() string
	CreateSnap() ([]byte, error)
	LoadSnap([]byte) error
}

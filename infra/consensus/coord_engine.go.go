package consensus

type CoordEngine interface {
	Handle(data []byte) error
	GenerateID() (uint64, error)

	GetSnapshot() ([]byte, error)
	ReloadSnapshot([]byte) error
}

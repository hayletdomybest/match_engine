package consensus

type CoordEngine interface {
	Handle(data []byte) error
	ReadHandle(index uint64, requestCtx []byte) error
	GenerateID() (uint64, error)
	CreateSyncRead(requestCtx []byte) chan uint64
	CancelSyncRead(requestCtx []byte)

	GetSnapshot() ([]byte, error)
	ReloadSnapshot([]byte) error
}

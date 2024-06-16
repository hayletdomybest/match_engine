package consensus

type ServerNode struct {
	NodeID uint64
	URL    string
}

type Server interface {
	Propose(data []byte) error
	AddNodes(nodes []ServerNode) error
}

package server

import (
	"fmt"
	"sync"
)

type ServerContext struct {
	NodeID uint64
	URL    string
	Peers  map[uint64]string

	mu       sync.Mutex
	sealed   bool
	sealedch chan struct{}
}

func NewContext(nodeID uint64, url string) *ServerContext {
	peers := make(map[uint64]string)
	peers[nodeID] = url

	return &ServerContext{
		NodeID: nodeID,
		URL:    url,

		Peers:    peers,
		sealed:   false,
		sealedch: make(chan struct{}),
	}
}

func (context *ServerContext) assertNotSealed() {
	context.mu.Lock()
	defer context.mu.Unlock()

	if context.sealed {
		panic(fmt.Errorf("app context has been sealed"))
	}
}

func (context *ServerContext) Sealed() *ServerContext {
	context.assertNotSealed()

	context.mu.Lock()
	defer context.mu.Unlock()
	context.sealed = true
	close(context.sealedch)
	return context
}

func (context *ServerContext) AppendPeer(nodeID uint64, nodeURL string) *ServerContext {
	context.assertNotSealed()

	context.mu.Lock()
	defer context.mu.Unlock()

	if nodeID == context.NodeID {
		context.URL = nodeURL
		return context
	}

	context.Peers[nodeID] = nodeURL
	return context
}

package cmd

import (
	"fmt"
	"sync"
)

type AppContext struct {
	NodeID uint64
	URL    string
	Peers  map[uint64]string

	mu       sync.Mutex
	sealed   bool
	sealedch chan struct{}
}

func NewContext(nodeID uint64, url string) *AppContext {
	peers := make(map[uint64]string)
	peers[nodeID] = url

	return &AppContext{
		NodeID: nodeID,
		URL:    url,

		Peers:    peers,
		sealed:   false,
		sealedch: make(chan struct{}),
	}
}

func (context *AppContext) assertNotSealed() {
	context.mu.Lock()
	defer context.mu.Unlock()

	if context.sealed {
		panic(fmt.Errorf("app context has been sealed"))
	}
}

func (context *AppContext) Sealed() *AppContext {
	context.assertNotSealed()

	context.mu.Lock()
	defer context.mu.Unlock()
	context.sealed = true
	close(context.sealedch)
	return context
}

func (context *AppContext) AppendPeer(nodeID uint64, nodeURL string) *AppContext {
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

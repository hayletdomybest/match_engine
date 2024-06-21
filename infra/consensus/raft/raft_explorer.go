package raft

import "match_engine/infra/consensus"

type RaftExplorer interface {
	GetNodes() ([]consensus.ServerNode, error)
	RegisterNode(nodeID uint64, url string) error
}

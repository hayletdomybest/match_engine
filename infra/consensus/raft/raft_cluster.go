package raft

import (
	"fmt"
	"match_engine/infra/consensus"
	"sync"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
)

// RaftCluster manage the node id and url
type RaftCluster struct {
	members map[uint64]string
	mutex   sync.RWMutex
}

// NewCluster create a Cluster from map
func NewCluster(peers map[uint64]string) *RaftCluster {
	return &RaftCluster{
		members: peers,
	}
}

// GetURL find the url
func (c *RaftCluster) GetURL(id uint64) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.members[id]
}

// AddMember add a new member to Cluster
func (c *RaftCluster) AddMember(id uint64, url string) {
	c.AddMembers(consensus.ServerNode{
		NodeID: id,
		URL:    url,
	})
}

func (c *RaftCluster) AddMembers(peers ...consensus.ServerNode) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, peer := range peers {
		c.members[peer.NodeID] = peer.URL
	}

}

// RemoveMember remove a existed member from Cluster
func (c *RaftCluster) RemoveMember(id uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.members, id)
}

func (c *RaftCluster) RemoveAllMember() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.members = make(map[uint64]string)
}

// HasMember check if the member in the Cluster
func (c *RaftCluster) HasMember(id uint64) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, ok := c.members[id]
	return ok
}

// ApplyConfigChange apply the Ready ConfChange Message
func (c *RaftCluster) ApplyConfigChange(cc raftpb.ConfChange) {
	switch cc.Type {
	case raftpb.ConfChangeAddNode, raftpb.ConfChangeAddLearnerNode:
		c.AddMember(cc.NodeID, string(cc.Context))
	case raftpb.ConfChangeRemoveNode:
		c.RemoveMember(cc.NodeID)
	}
}

func (c *RaftCluster) GetPeers() []raft.Peer {
	var peers []raft.Peer
	for id, url := range c.members {
		peers = append(peers, raft.Peer{
			ID:      id,
			Context: []byte(url),
		})
	}
	return peers
}

func (c *RaftCluster) CheckNodeUrl(id uint64, testUrl string) error {
	url := c.GetURL(id)
	if len(url) == 0 {
		return fmt.Errorf("node %d url is empty", id)
	}

	if url != testUrl {
		return fmt.Errorf("node %d url is not same as peer %d url", id, id)
	}

	return nil
}

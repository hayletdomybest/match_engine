package server

import (
	"match_engine/app/cmd/common"
	"match_engine/infra/consensus"
	"match_engine/infra/consensus/raft"
	serviceexplore "match_engine/infra/service_explore"
	"match_engine/utils"
)

type ServerExplorer struct {
	explorer serviceexplore.ServiceExplore
}

var _ raft.RaftExplorer = (*ServerExplorer)(nil)

func NewServerExplorer(explorer serviceexplore.ServiceExplore) *ServerExplorer {
	return &ServerExplorer{
		explorer: explorer,
	}
}

func (s *ServerExplorer) GetNodes() ([]consensus.ServerNode, error) {
	w, err := s.explorer.GetWatcher(common.NodeExplorePath)
	if err != nil {
		return []consensus.ServerNode{}, err
	}
	service, err := w.GetServices()
	if err != nil {
		return []consensus.ServerNode{}, err
	}
	if len(service) == 0 {
		return []consensus.ServerNode{}, nil
	}

	nodes := utils.Select(service, func(pair serviceexplore.ServiceWatchKeyValPair) consensus.ServerNode {
		nodeID, _ := common.ParseNodeId(pair.Key)
		return consensus.ServerNode{
			NodeID: nodeID,
			URL:    pair.Val,
		}
	})

	return nodes, nil
}

func (s *ServerExplorer) RegisterNode(nodeID uint64, url string) error {
	_, err := s.explorer.Register(common.GetPath(nodeID), url, 5)
	return err
}

func (s *ServerExplorer) Close() error {
	return s.explorer.Close()
}

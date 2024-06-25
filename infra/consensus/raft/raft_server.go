package raft

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"match_engine/infra/consensus"
	"match_engine/infra/log"
	"match_engine/utils"
	"path/filepath"

	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/google/uuid"
)

type RaftServer struct {
	ctx    context.Context
	logger log.Logger
	ticker *time.Ticker
	mu     sync.Mutex

	nodeID        uint64
	url           string
	join          bool
	heartbeatTick int
	electionTick  int
	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	raft    raft.Node
	storage *raft.MemoryStorage

	snapshotter *RaftSnapshotter
	walPath     string
	wal         *wal.WAL

	cluster    *RaftCluster
	explorer   RaftExplorer
	transport  *rafthttp.Transport
	httpServer *http.Server
	tls        *HttpTransportTLS

	configChangeC chan raftpb.ConfChange
	proposeC      chan []byte
	readStateC    chan raft.ReadState
	stopC         chan struct{}
	doneC         chan struct{}
	errorC        chan error

	engine consensus.CoordEngine
}

var _ consensus.Server = (*RaftServer)(nil)

func NewRaftServer(conf *RaftServerConf) *RaftServer {
	storage := raft.NewMemoryStorage()
	id := conf.NodeID
	if len(conf.HomePath) == 0 {
		panic("raft home path can not be empty")
	}

	server := &RaftServer{
		nodeID: id,
		url:    conf.URL,
		join:   conf.Join,
		ctx:    conf.Context,
		logger: conf.Logger,
		ticker: conf.Ticker,

		heartbeatTick: conf.HeartbeatTick,
		electionTick:  conf.ElectionTick,
		storage:       storage,
		snapshotter: NewRaftSnapshot(
			filepath.Join(conf.HomePath, fmt.Sprintf("%s-%d", DefaultSnapshotDir, id)),
			conf.SnapshotTick,
		),
		walPath: filepath.Join(conf.HomePath, fmt.Sprintf("%s-%d", DefaultWalDir, id)),

		cluster: NewCluster(conf.Peers),

		tls:           conf.TLS,
		configChangeC: make(chan raftpb.ConfChange),
		proposeC:      make(chan []byte),
		readStateC:    make(chan raft.ReadState),
		stopC:         make(chan struct{}),
		doneC:         make(chan struct{}),
		errorC:        make(chan error),

		engine:   conf.Engine,
		explorer: conf.Explorer,
	}
	server.snapshotter.
		SetTrigger(func(_ *snap.Snapshotter) {
			server.triggerSnap()
		}).
		SetErrorC(server.errorC)

	return server
}

func (srv *RaftServer) Start() error {
	raftConf := &raft.Config{
		ID:              srv.nodeID,
		ElectionTick:    srv.electionTick,
		HeartbeatTick:   srv.heartbeatTick,
		Storage:         srv.storage,
		MaxSizePerMsg:   math.MaxUint16,
		MaxInflightMsgs: 256,
	}

	isRestart, err := srv.replayWAL()
	if err != nil {
		return err
	}

	if srv.explorer != nil {
		srv.join = true
	}

	if isRestart || srv.join {
		srv.raft = raft.RestartNode(raftConf)
	} else {
		srv.raft = raft.StartNode(raftConf, srv.cluster.GetPeers())
	}

	snap, err := srv.storage.Snapshot()
	if err != nil {
		return err
	}
	srv.confState = snap.Metadata.ConfState
	srv.snapshotIndex = snap.Metadata.Index
	srv.appliedIndex = snap.Metadata.Index

	srv.ResetNode()

	if err := srv.httpTransportStart(); err != nil {
		return err
	}

	if srv.explorer != nil {
		err := srv.explorer.RegisterNode(srv.nodeID, srv.url)
		if err != nil {
			return err
		}
		nodes, err := srv.explorer.GetNodes()
		if err != nil {
			return err
		}

		srv.cluster.RemoveAllMember()
		srv.SyncMembers(nodes...)
		if err := srv.cluster.CheckNodeUrl(srv.nodeID, srv.url); err != nil {
			return err
		}

		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-srv.doneC:
					return
				case <-ticker.C:
					nodes, err := srv.explorer.GetNodes()
					if err != nil {
						srv.errorC <- err
						continue
					}
					srv.SyncMembers(nodes...)
				}
			}
		}()
	}

	srv.serveProposeChannels()
	srv.serveRaftHandlerChannels()
	srv.serveRaftRead()

	return nil
}

func (srv *RaftServer) Stop() {
	if srv.stopC == nil {
		return
	}
	defer srv.mu.Unlock()
	srv.mu.Lock()
	close(srv.stopC)
	srv.stopC = nil
	srv.raft.Stop()
	srv.wal.Close()
	srv.httpServer.Close()
	srv.transport.Stop()

	close(srv.doneC)
}

func (srv *RaftServer) IsReady() bool {
	// TODO
	return true
}

func (srv *RaftServer) Propose(data []byte) error {
	srv.proposeC <- data
	return nil
}

func (srv *RaftServer) ReadIndex() (<-chan uint64, error) {
	rtcx := []byte(uuid.New().String())
	res := srv.engine.CreateSyncRead(rtcx)
	if err := srv.raft.ReadIndex(srv.ctx, rtcx); err != nil {
		srv.engine.CancelSyncRead(rtcx)
		return nil, err
	}
	return res, nil
}

func (srv *RaftServer) AddNodes(nodes ...consensus.ServerNode) error {
	for _, node := range nodes {
		url := node.URL
		id := node.NodeID
		go func() {
			select {
			case <-time.After(10 * time.Second):
				srv.errorC <- errors.New("add node timeout")
			case srv.configChangeC <- raftpb.ConfChange{
				Context: []byte(url),
				Type:    raftpb.ConfChangeAddNode,
				NodeID:  id,
			}:
			}
		}()
	}
	return nil
}

func (srv *RaftServer) RemoveNodes(nodeIDs ...uint64) error {
	for _, nodeID := range nodeIDs {
		if !srv.cluster.HasMember(nodeID) {
			continue
		}
		go func(id uint64) {
			select {
			case <-time.After(10 * time.Second):
				srv.errorC <- errors.New("remove node timeout")
			case srv.configChangeC <- raftpb.ConfChange{
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: id,
			}:
			}
		}(nodeID)

	}
	return nil
}

func (srv *RaftServer) applySoftSnap(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	if snapshotToSave.Metadata.Index <= srv.appliedIndex {
		srv.errorC <- fmt.Errorf(
			"fail: apply snapshot index [%d] should > progress.appliedIndex [%d]",
			snapshotToSave.Metadata.Index, srv.appliedIndex)
		return
	}
	srv.confState = snapshotToSave.Metadata.ConfState
	srv.snapshotIndex = snapshotToSave.Metadata.Index
	srv.appliedIndex = snapshotToSave.Metadata.Index
}

func (srv *RaftServer) applyHardSnap(snap raftpb.Snapshot) error {
	// save the snapshot file before writing the snapshot to the wal.
	// This makes it possible for the snapshot file to become orphaned, but prevents
	// a WAL snapshot entry from having no corresponding snapshot file.
	if err := srv.snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}
	if err := srv.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	return srv.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (srv *RaftServer) triggerSnap() {
	if srv.appliedIndex == srv.snapshotIndex {
		return
	}

	data, err := srv.engine.GetSnapshot()
	if err != nil {
		srv.errorC <- fmt.Errorf("trigger snapshot error:%s", err.Error())
		return
	}
	if len(data) == 0 {
		return
	}

	snap, err := srv.storage.CreateSnapshot(srv.appliedIndex, &srv.confState, data)
	if err != nil {
		srv.errorC <- fmt.Errorf("trigger snapshot error:%s", err.Error())
		return
	}

	if err := srv.applyHardSnap(snap); err != nil {
		srv.errorC <- fmt.Errorf("trigger snapshot error:%s", err.Error())
		return
	}

	var compactIndex uint64
	if srv.appliedIndex > uint64(SnapshotCatchUpEntriesN) {
		compactIndex = srv.appliedIndex - uint64(SnapshotCatchUpEntriesN)
	}
	if err := srv.storage.Compact(compactIndex); err != nil {
		if err != raft.ErrCompacted {
			srv.errorC <- err
			return
		}
	}
	srv.snapshotIndex = srv.appliedIndex
}

func (rc *RaftServer) GetLeader() uint64 {
	return rc.raft.Status().Lead
}

func (rc *RaftServer) GetId() uint64 {
	return rc.nodeID
}

func (rc *RaftServer) IsLeader() bool {
	return rc.GetLeader() == rc.nodeID
}

func (rc *RaftServer) SyncMembers(nodes ...consensus.ServerNode) {
	var m = utils.SliceToMap(nodes, func(n consensus.ServerNode) uint64 { return n.NodeID })

	var addNode []consensus.ServerNode

	for _, node := range nodes {
		if rc.cluster.HasMember(node.NodeID) {
			continue
		}
		rc.cluster.AddMember(node.NodeID, node.URL)
		rc.transport.AddPeer(types.ID(node.NodeID), []string{node.URL})
		id, _ := rc.engine.GenerateID()
		rc.raft.ApplyConfChange(raftpb.ConfChange{
			ID:      id,
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  node.NodeID,
			Context: []byte(node.URL),
		})
		addNode = append(addNode, consensus.ServerNode{
			NodeID: node.NodeID,
			URL:    node.URL,
		})
	}
	var removeNodeIDs []uint64
	for nodeID := range rc.cluster.members {
		if _, existed := m[nodeID]; !existed {

			if nodeID == rc.nodeID {
				rc.errorC <- fmt.Errorf("node%d disconnect with server discovered", nodeID)
				rc.Stop()
			}
			rc.cluster.RemoveMember(nodeID)
			rc.transport.RemovePeer(types.ID(nodeID))
			id, _ := rc.engine.GenerateID()
			rc.raft.ApplyConfChange(raftpb.ConfChange{
				ID:     id,
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: nodeID,
			})
			removeNodeIDs = append(removeNodeIDs, nodeID)
		}
	}
	rc.AddNodes(addNode...)
	rc.RemoveNodes(removeNodeIDs...)
}

func (rc *RaftServer) CatchError() <-chan error {
	return rc.errorC
}

func (rc *RaftServer) Done() <-chan struct{} {
	return rc.doneC
}

func (rc *RaftServer) ResetNode() {
	for _, nodeID := range rc.confState.Nodes {
		rc.raft.ApplyConfChange(raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: nodeID,
		})
	}
}

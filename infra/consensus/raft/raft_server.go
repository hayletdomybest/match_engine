package raft

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"match_engine/infra/consensus"
	"match_engine/infra/log"
	"path/filepath"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
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
	transport  *rafthttp.Transport
	httpServer *http.Server
	tls        *HttpTransportTLS

	configChangeC chan raftpb.ConfChange
	proposeC      chan []byte
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
		stopC:         make(chan struct{}),
		doneC:         make(chan struct{}),
		errorC:        make(chan error),

		engine: conf.Engine,
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

	if err := srv.cluster.CheckNodeUrl(srv.nodeID, srv.url); err != nil {
		return err
	}

	isRestart, err := srv.replayWAL()
	if err != nil {
		return err
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

	if err := srv.httpTransportStart(); err != nil {
		return err
	}

	srv.serveProposeChannels()
	srv.serveRaftHandlerChannels()

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

func (srv *RaftServer) Propose(data []byte) error {
	srv.proposeC <- data
	return nil
}

func (srv *RaftServer) AddNodes(nodes ...consensus.ServerNode) error {
	for _, node := range nodes {

		srv.configChangeC <- raftpb.ConfChange{
			Context: []byte(node.URL),
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  node.NodeID,
		}
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

func (rc *RaftServer) CatchError() <-chan error {
	return rc.errorC
}

func (rc *RaftServer) Done() <-chan struct{} {
	return rc.doneC
}

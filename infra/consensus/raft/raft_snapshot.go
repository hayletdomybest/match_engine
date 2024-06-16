package raft

import (
	"match_engine/utils"
	"sync"

	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal/walpb"
)

type RaftSnapshotter struct {
	snapshotter  *snap.Snapshotter
	snapshotTick int
	curTick      int

	trigger func(snapshotter *snap.Snapshotter)

	errorC chan<- error
	mu     sync.Mutex
}

func NewRaftSnapshot(
	path string,
	snapshotTick int) *RaftSnapshotter {
	if snapshotTick == 0 {
		panic("can not set zero on snapshotTick")
	}
	snapshot := &RaftSnapshotter{
		snapshotTick: snapshotTick,
		curTick:      0,
	}

	utils.MkdirAll(path)
	snapshot.snapshotter = snap.New(path)

	return snapshot
}

func (snapshot *RaftSnapshotter) Tick() {
	if snapshot.trigger == nil {
		panic("snapshot trigger can not be nil")
	}
	snapshot.curTick++

	if snapshot.curTick != snapshot.snapshotTick {
		return
	}

	if !snapshot.mu.TryLock() {
		snapshot.curTick--
		return
	}

	go func() {
		snapshot.trigger(snapshot.snapshotter)
		snapshot.mu.Unlock()
		snapshot.curTick = 0
	}()
}

func (snapshot *RaftSnapshotter) SetTrigger(trigger func(*snap.Snapshotter)) *RaftSnapshotter {
	snapshot.trigger = trigger
	return snapshot
}

func (snapshot *RaftSnapshotter) SetErrorC(errorC chan<- error) *RaftSnapshotter {
	snapshot.errorC = errorC
	return snapshot
}

func (snapshot *RaftSnapshotter) LoadNewestAvailable(walSnaps []walpb.Snapshot) (*raftpb.Snapshot, error) {
	return snapshot.snapshotter.LoadNewestAvailable(walSnaps)
}

func (snapshot *RaftSnapshotter) SaveSnap(snap raftpb.Snapshot) error {
	return snapshot.snapshotter.SaveSnap(snap)
}

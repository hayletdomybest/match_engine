package raft

import (
	"match_engine/utils"

	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
)

func (srv *RaftServer) replayWAL() (bool, error) {
	isRestart := wal.Exist(srv.walPath)

	utils.MkdirAll(srv.walPath)
	walSnaps, err := wal.ValidSnapshotEntries(srv.walPath)
	if err != nil {
		return isRestart, err
	}
	snapshot, err := srv.snapshotter.LoadNewestAvailable(walSnaps)
	if err != nil && err != snap.ErrNoSnapshot {
		return isRestart, err
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		srv.engine.ReloadSnapshot(snapshot.Data)
		srv.storage.ApplySnapshot(*snapshot)
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	w, err := wal.Open(srv.walPath, walsnap)
	if err != nil {
		return isRestart, err
	}
	srv.wal = w

	_, st, ents, err := srv.wal.ReadAll()
	if err != nil {
		return isRestart, err
	}
	srv.storage.SetHardState(st)
	srv.storage.Append(ents)
	return isRestart, nil
}

package raft

import (
	"match_engine/utils"

	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/pkg/errors"
)

func (srv *RaftServer) replayWAL() (bool, error) {
	existed := wal.Exist(srv.walPath)
	utils.MkdirAll(srv.walPath)

	if !existed {
		w, err := wal.Create(srv.walPath, nil)
		if err != nil {
			return existed, errors.Errorf("create wal error (%v)", err)
		}
		w.Close()
	}
	walSnaps, err := wal.ValidSnapshotEntries(srv.walPath)
	if err != nil {
		return existed, err
	}
	snapshot, err := srv.snapshotter.LoadNewestAvailable(walSnaps)
	if err != nil && err != snap.ErrNoSnapshot {
		return existed, err
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		srv.engine.ReloadSnapshot(snapshot.Data)
		srv.storage.ApplySnapshot(*snapshot)
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	w, err := wal.Open(srv.walPath, walsnap)
	if err != nil {
		return existed, err
	}
	srv.wal = w

	_, st, ents, err := srv.wal.ReadAll()
	if err != nil {
		return existed, err
	}
	srv.storage.SetHardState(st)
	srv.storage.Append(ents)
	return existed, nil
}

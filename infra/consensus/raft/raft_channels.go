package raft

import (
	"errors"

	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
)

func (srv *RaftServer) serveProposeChannels() {
	go func() {
		defer close(srv.proposeC)
		defer close(srv.configChangeC)

		for {
			select {
			case <-srv.stopC:
				return
			case prop, ok := <-srv.proposeC:
				if !ok {
					srv.proposeC = nil
				} else {
					srv.raft.Propose(srv.ctx, prop)
				}
			case cc, ok := <-srv.configChangeC:
				if !ok {
					srv.configChangeC = nil
				} else {
					id, err := srv.engine.GenerateID()
					if err != nil {
						srv.errorC <- err
						return
					}
					cc.ID = id
					srv.raft.ProposeConfChange(srv.ctx, cc)
				}
			}
		}
	}()
}

func (srv *RaftServer) serveRaftHandlerChannels() {
	go func() {
		defer srv.ticker.Stop()
		for {
			select {
			case <-srv.stopC:
				return
			case <-srv.ticker.C:
				srv.raft.Tick()
				srv.snapshotter.Tick()
			case rd := <-srv.raft.Ready():
				srv.mu.Lock()
				if !raft.IsEmptySnap(rd.Snapshot) {
					srv.logger.Info("node %d has snapshot at index %d term %d\n", srv.nodeID, rd.Snapshot.Metadata.Index, rd.Snapshot.Metadata.Term)
				}

				if !raft.IsEmptySnap(rd.Snapshot) {
					srv.applyHardSnap(rd.Snapshot)
				}
				srv.wal.Save(rd.HardState, rd.Entries)
				if !raft.IsEmptySnap(rd.Snapshot) {
					srv.engine.ReloadSnapshot(rd.Snapshot.Data)
					srv.storage.ApplySnapshot(rd.Snapshot)
					srv.applySoftSnap(rd.Snapshot)
				}
				srv.storage.Append(rd.Entries)
				srv.transport.Send(srv.processMessages(rd.Messages))
				srv.publishEntries(rd.CommittedEntries)
				srv.raft.Advance()
				srv.mu.Unlock()
			}
		}
	}()
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (srv *RaftServer) publishEntries(ents []raftpb.Entry) {
	if len(ents) == 0 {
		return
	}

	for i, ent := range ents {
		if ent.Index <= srv.appliedIndex {
			srv.logger.Info("node %d skip entry index %d because node has applied index %d", srv.nodeID, ent.Index, srv.appliedIndex)
			continue
		}

		switch ent.Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				break
			}
			srv.engine.Handle(ent.Data)
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)

			srv.confState = *srv.raft.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				srv.logger.Info("node %d applied ConfChangeAddNode", srv.nodeID)
				if len(cc.Context) > 0 && cc.NodeID != srv.nodeID {
					srv.cluster.AddMember(cc.NodeID, string(cc.Context))
					srv.logger.Info("node%d add node%d addr %s", srv.nodeID, cc.NodeID, string(cc.Context))
					srv.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(srv.nodeID) {
					srv.errorC <- errors.New("node has been removed from the cluster! shutting down")
					srv.Stop()
					return
				}
				if srv.cluster.HasMember(cc.NodeID) {
					srv.cluster.RemoveMember(cc.NodeID)
					srv.transport.RemovePeer(types.ID(cc.NodeID))
				}
			}
		}
	}

	select {
	case <-srv.stopC:
		return
	default:
		srv.appliedIndex = ents[len(ents)-1].Index
	}
}

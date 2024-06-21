package raft

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/pkg/errors"
)

func (srv *RaftServer) httpTransportStart() error {
	srv.transport = &rafthttp.Transport{
		DialTimeout: 5 * time.Second,
		ID:          types.ID(srv.nodeID),
		ClusterID:   1,
		Raft:        srv,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(int(srv.nodeID))),
		ErrorC:      srv.errorC,
	}
	if err := srv.transport.Start(); err != nil {
		return errors.Errorf("Failed transport start (%v)", err)
	}

	u, err := url.Parse(srv.url)
	if err != nil {
		return errors.Errorf("Failed parsing URL (%v)", err)
	}
	for id, member := range srv.cluster.members {
		srv.transport.AddPeer(types.ID(id), []string{member})
	}

	var ln net.Listener
	if srv.tls != nil && srv.tls.Enable {
		cert, err := tls.LoadX509KeyPair(srv.tls.Cert, srv.tls.Key)
		if err != nil {
			return errors.Errorf("Failed loading cert (%v)", err)
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		ln, err = tls.Listen("tcp", u.Host, tlsConfig)
		if err != nil {
			return errors.Errorf("Failed listening (%v)", err)
		}
	} else {
		ln, err = net.Listen("tcp", u.Host)
		if err != nil {
			errors.Errorf("Failed listening (%v)", err)
		}
	}

	srv.httpServer = &http.Server{Handler: srv.transport.Handler()}

	go func() {
		defer srv.httpServer.Close()
		srv.logger.Info("node%d is listening on addr %s", srv.nodeID, ln.Addr().String())
		err = srv.httpServer.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			srv.errorC <- errors.Errorf("Http server close (%v)", err)
		}
	}()

	return nil
}

func (srv *RaftServer) Process(ctx context.Context, m raftpb.Message) error {
	return srv.raft.Step(ctx, m)
}

func (srv *RaftServer) IsIDRemoved(id uint64) bool {
	return !srv.cluster.HasMember(id)
}

func (srv *RaftServer) ReportUnreachable(id uint64) {
	srv.raft.ReportUnreachable(id)
}

func (srv *RaftServer) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	srv.raft.ReportSnapshot(id, status)
}

func (srv *RaftServer) processMessages(ms []raftpb.Message) []raftpb.Message {
	for i := 0; i < len(ms); i++ {
		if ms[i].Type == raftpb.MsgSnap {
			ms[i].Snapshot.Metadata.ConfState = srv.confState
		}
	}
	return ms
}

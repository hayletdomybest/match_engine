package raft

import (
	"context"
	"match_engine/infra/consensus"
	"match_engine/infra/log"
	"time"
)

type HttpTransportTLS struct {
	Enable bool
	Key    string
	Cert   string
}

type RaftServerConf struct {
	NodeID   uint64
	URL      string
	Peers    map[uint64]string
	HomePath string

	Ticker        *time.Ticker
	ElectionTick  int
	HeartbeatTick int
	SnapshotTick  int

	Logger log.Logger
	TLS    *HttpTransportTLS

	Context context.Context
	Engine  consensus.CoordEngine
}

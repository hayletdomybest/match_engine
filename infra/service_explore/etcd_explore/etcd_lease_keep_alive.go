package etcdexplore

import (
	"context"
	serviceexplore "match_engine/infra/service_explore"

	"go.etcd.io/etcd/clientv3"
)

type EtcdLeaseKeepALiveConf struct {
	Ctx           context.Context
	Client        *clientv3.Client //etcd client
	LeaseID       clientv3.LeaseID
	KeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	Key           string
	Value         string
}

type EtcdLeaseKeepALiveFn func(*clientv3.LeaseKeepAliveResponse)

type EtcdLeaseKeepALive struct {
	ctx           context.Context
	cli           *clientv3.Client //etcd client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	value         string

	subscribes []EtcdLeaseKeepALiveFn
}

var _ serviceexplore.ServiceKeepAlive = (*EtcdLeaseKeepALive)(nil)

func NewEtcdLeaseKeepALive(conf *EtcdLeaseKeepALiveConf) *EtcdLeaseKeepALive {
	return &EtcdLeaseKeepALive{
		ctx:           conf.Ctx,
		cli:           conf.Client,
		leaseID:       conf.LeaseID,
		keepAliveChan: conf.KeepAliveChan,
		key:           conf.Key,
		value:         conf.Value,
	}
}

func (alive *EtcdLeaseKeepALive) Close() error {
	return alive.cli.Close()
}

func (alive *EtcdLeaseKeepALive) handle(repo *clientv3.LeaseKeepAliveResponse) {
	for _, subscribe := range alive.subscribes {
		go subscribe(repo)
	}
}

func (alive *EtcdLeaseKeepALive) Subscribe(fn EtcdLeaseKeepALiveFn) {
	alive.subscribes = append(alive.subscribes, fn)
}

func (alive *EtcdLeaseKeepALive) Keep() {
	go func() {
		for leaseKeepResp := range alive.keepAliveChan {
			alive.handle(leaseKeepResp)
		}
	}()
}

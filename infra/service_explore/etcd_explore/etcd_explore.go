package etcdexplore

import (
	serviceexplore "match_engine/infra/service_explore"
	"sync"

	"go.etcd.io/etcd/clientv3"
)

type EtcdExploreConf clientv3.Config

type EtcdExplore struct {
	keepAlives map[string]serviceexplore.ServiceKeepAlive
	watchers   map[string]serviceexplore.ServiceWatcher
	conf       *EtcdExploreConf
}

var _ serviceexplore.ServiceExplore = (*EtcdExplore)(nil)

func NewEtcdExplore(conf *EtcdExploreConf) *EtcdExplore {
	return &EtcdExplore{
		keepAlives: make(map[string]serviceexplore.ServiceKeepAlive),
		watchers:   make(map[string]serviceexplore.ServiceWatcher),
		conf:       conf,
	}
}

func (explore *EtcdExplore) Close() error {
	var wg sync.WaitGroup
	var err error = nil
	for _, alive := range explore.keepAlives {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_e := alive.Close()
			if _e != nil && err == nil {
				err = _e
			}
		}()
	}

	for _, watch := range explore.watchers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_e := watch.Close()
			if _e != nil && err == nil {
				err = _e
			}
		}()
	}

	wg.Wait()
	return err
}

func (explore *EtcdExplore) GetWatcher(prefix string) (serviceexplore.ServiceWatcher, error) {
	if _, existed := explore.watchers[prefix]; existed {
		return explore.watchers[prefix], nil
	}

	conf := explore.conf
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: conf.DialTimeout,
	})
	if err != nil {
		return nil, err
	}
	watcher := NewEtcdWatcher(&EtcdWatcherConf{
		Ctx:    conf.Context,
		Client: cli,
		Prefix: prefix,
	})
	explore.watchers[prefix] = watcher
	return watcher, nil
}

func (explore *EtcdExplore) Register(key string, val string, ttl uint64) (serviceexplore.ServiceKeepAlive, error) {
	if _, existed := explore.keepAlives[key]; existed {
		return explore.keepAlives[key], nil
	}

	conf := explore.conf
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: conf.DialTimeout,
	})
	if err != nil {
		return nil, err
	}

	resp, err := cli.Grant(conf.Context, int64(ttl))
	if err != nil {
		return nil, err
	}
	_, err = cli.Put(conf.Context, key, val, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}

	keepAliveChan, err := cli.KeepAlive(conf.Context, resp.ID)

	if err != nil {
		return nil, err
	}
	leaseID := resp.ID

	keepAlive := NewEtcdLeaseKeepALive(&EtcdLeaseKeepALiveConf{
		Ctx:           conf.Context,
		Client:        cli,
		Key:           key,
		Value:         val,
		KeepAliveChan: keepAliveChan,
		LeaseID:       leaseID,
	})

	explore.keepAlives[key] = keepAlive

	return keepAlive, nil
}

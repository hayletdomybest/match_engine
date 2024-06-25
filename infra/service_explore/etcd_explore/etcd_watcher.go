package etcdexplore

import (
	"context"
	"errors"
	serviceexplore "match_engine/infra/service_explore"
	"sync"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type EtcdWatcherConf struct {
	Ctx    context.Context
	Client *clientv3.Client
	Prefix string
}

type EtcdWatcher struct {
	events map[serviceexplore.ServiceWatcherEvent][]serviceexplore.ServiceWatcherFn
	ctx    context.Context
	cli    *clientv3.Client
	prefix string

	mu sync.Mutex
}

var _ serviceexplore.ServiceWatcher = (*EtcdWatcher)(nil)

func NewEtcdWatcher(conf *EtcdWatcherConf) *EtcdWatcher {
	watcher := &EtcdWatcher{
		ctx:    conf.Ctx,
		cli:    conf.Client,
		prefix: conf.Prefix,
		events: make(map[serviceexplore.ServiceWatcherEvent][]serviceexplore.ServiceWatcherFn),
	}
	watcher.events[serviceexplore.AddNode] = make([]serviceexplore.ServiceWatcherFn, 0)
	watcher.events[serviceexplore.DelNode] = make([]serviceexplore.ServiceWatcherFn, 0)

	return watcher
}

// Close implements serviceexplore.ServiceWatcher.
func (watcher *EtcdWatcher) Close() error {
	return watcher.cli.Close()
}

// Subscribe implements serviceexplore.ServiceWatcher.
func (watcher *EtcdWatcher) Subscribe(event serviceexplore.ServiceWatcherEvent, fn func(key string, val string)) {
	watcher.events[event] = append(watcher.events[event], fn)
}

func (watcher *EtcdWatcher) GetServices() ([]serviceexplore.ServiceWatchKeyValPair, error) {
	resp, err := watcher.cli.Get(watcher.ctx, watcher.prefix, clientv3.WithPrefix())
	if err != nil {
		return []serviceexplore.ServiceWatchKeyValPair{}, err
	}

	var res []serviceexplore.ServiceWatchKeyValPair

	for _, ev := range resp.Kvs {
		res = append(res, serviceexplore.ServiceWatchKeyValPair{
			Key: string(ev.Key),
			Val: string(ev.Value),
		})
	}
	return res, nil
}

// Watch implements serviceexplore.ServiceWatcher.
func (watcher *EtcdWatcher) Watch() error {
	if !watcher.mu.TryLock() {
		return errors.New("watcher is watching")
	}
	defer watcher.mu.Unlock()
	go func() {
		watchC := watcher.cli.Watch(watcher.ctx, watcher.prefix, clientv3.WithPrefix())
		for wresp := range watchC {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					watcher.handle(serviceexplore.AddNode, string(ev.Kv.Key), string(ev.Kv.Value))
				case mvccpb.DELETE:
					watcher.handle(serviceexplore.DelNode, string(ev.Kv.Key), string(ev.Kv.Value))
				}
			}
		}
	}()

	return nil
}

func (watcher *EtcdWatcher) handle(event serviceexplore.ServiceWatcherEvent, key string, val string) {
	for _, fn := range watcher.events[event] {
		_fn := fn
		go func() {
			_fn(key, val)
		}()
	}
}

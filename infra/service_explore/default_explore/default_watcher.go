package defaultexplore

import serviceexplore "match_engine/infra/service_explore"

type DefaultWatcher struct {
	prefix string
	pairs  []serviceexplore.ServiceWatchKeyValPair
	events map[serviceexplore.ServiceWatcherEvent][]serviceexplore.ServiceWatcherFn
}

var _ serviceexplore.ServiceWatcher = (*DefaultWatcher)(nil)

func NewDefaultWatcher(prefix string) *DefaultWatcher {
	watcher := &DefaultWatcher{
		prefix: prefix,
		events: make(map[serviceexplore.ServiceWatcherEvent][]serviceexplore.ServiceWatcherFn),
	}

	watcher.events[serviceexplore.AddNode] = make([]serviceexplore.ServiceWatcherFn, 0)
	watcher.events[serviceexplore.DelNode] = make([]serviceexplore.ServiceWatcherFn, 0)
	return watcher
}

func (d *DefaultWatcher) Close() error {
	return nil
}

func (d *DefaultWatcher) SetPairs(pairs ...serviceexplore.ServiceWatchKeyValPair) error {
	d.pairs = pairs
	return nil
}

func (d *DefaultWatcher) GetServices() ([]serviceexplore.ServiceWatchKeyValPair, error) {
	return d.pairs, nil
}

func (d *DefaultWatcher) Subscribe(event serviceexplore.ServiceWatcherEvent, fn func(key string, val string)) {
	d.events[event] = append(d.events[event], fn)
}

func (d *DefaultWatcher) Watch() error {
	for _, pair := range d.pairs {
		for _, fn := range d.events[serviceexplore.AddNode] {
			go fn(pair.Key, pair.Val)
		}
	}

	return nil
}

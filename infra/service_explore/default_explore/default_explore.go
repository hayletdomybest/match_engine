package defaultexplore

import serviceexplore "match_engine/infra/service_explore"

type DefaultExplore struct {
	watcher map[string]*DefaultWatcher
}

var _ serviceexplore.ServiceExplore = (*DefaultExplore)(nil)

func NewDefaultExplore() *DefaultExplore {
	return &DefaultExplore{
		watcher: make(map[string]*DefaultWatcher),
	}
}

// Close implements serviceexplore.ServiceExplore.
func (d *DefaultExplore) Close() error {
	return nil
}

// GetWatcher implements serviceexplore.ServiceExplore.
func (d *DefaultExplore) GetWatcher(prefix string) (serviceexplore.ServiceWatcher, error) {
	watcher, err := d.GetDefaultWatcher(prefix)
	return watcher, err
}

func (d *DefaultExplore) GetDefaultWatcher(prefix string) (*DefaultWatcher, error) {
	if _, existed := d.watcher[prefix]; existed {
		return d.watcher[prefix], nil
	}

	watcher := NewDefaultWatcher(prefix)
	d.watcher[prefix] = watcher

	return watcher, nil
}

// Register implements serviceexplore.ServiceExplore.
func (d *DefaultExplore) Register(key string, val string, ttl uint64) (serviceexplore.ServiceKeepAlive, error) {
	// do nothing
	return &DefaultLeaseKeepAlive{}, nil
}

package defaultexplore

import serviceexplore "match_engine/infra/service_explore"

type DefaultLeaseKeepAlive struct {
}

var _ serviceexplore.ServiceKeepAlive = (*DefaultLeaseKeepAlive)(nil)

func NewDefaultLeaseKeepAlive() *DefaultLeaseKeepAlive {
	return &DefaultLeaseKeepAlive{}
}

// Close implements serviceexplore.ServiceExplore.
func (d *DefaultLeaseKeepAlive) Close() error {
	return nil
}

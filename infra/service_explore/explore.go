package serviceexplore

type ServiceWatcherEvent int

type ServiceWatcherFn = func(key string, val string)

type ServiceWatchKeyValPair struct {
	Key string
	Val string
}

const (
	AddNode = iota
	DelNode
)

type ServiceWatcher interface {
	Subscribe(event ServiceWatcherEvent, fn ServiceWatcherFn)
	GetServices() ([]ServiceWatchKeyValPair, error)
	Watch() error
	Close() error
}

type ServiceKeepAlive interface {
	Close() error
}

type ServiceExplore interface {
	GetWatcher(prefix string) (ServiceWatcher, error)
	// ttl second
	Register(key string, val string, ttl uint64) (ServiceKeepAlive, error)
	Close() error
}

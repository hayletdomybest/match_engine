package server

import (
	"context"
	"fmt"
	"match_engine/app/cmd/common"
	"match_engine/app/explorer"
	"match_engine/app/helloworld"
	"match_engine/infra/consensus/raft"
	"match_engine/infra/db"
	"match_engine/infra/log"
	serviceexplore "match_engine/infra/service_explore"
	defaultexplore "match_engine/infra/service_explore/default_explore"
	etcdexplore "match_engine/infra/service_explore/etcd_explore"
	"match_engine/utils"
	"path"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

var services []interface{}

func inject(c *dig.Container, srv interface{}) {
	services = append(services, srv)
	c.Provide(srv)
}

func sealed(c *dig.Container) error {
	for _, srv := range services {
		err := c.Invoke(srv)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFromContainer[T any](c *dig.Container, srv T) (T, error) {
	var res T
	err := c.Invoke(func(item T) {
		res = item
	})
	if err != nil {
		return res, err
	}

	t := reflect.TypeOf(srv)
	if !reflect.ValueOf(res).IsValid() {
		return res, fmt.Errorf("can not invoke resource %s", t.Name())
	}

	return res, nil
}

func initContainer(ctx *common.AppContext) *dig.Container {
	c := dig.New()
	gin.SetMode(ctx.Mode)
	r := gin.Default()
	group := r.Group(common.ApiRootPath)

	inject(c, func() *gin.Engine {
		return r
	})
	inject(c, func() gin.IRoutes {
		return group
	})

	inject(c, func() *common.AppContext {
		return ctx
	})

	inject(c, NewServer)

	//repository
	inject(c, db.NewHelloWorldKv)
	inject(c, db.NewInMemoryDBContext)

	//controller
	inject(c, helloworld.NewHelloWorldController)
	inject(c, explorer.NewExplorerController)

	// explore
	inject(c, func(ctx *common.AppContext) serviceexplore.ServiceExplore {
		if len(ctx.AppConfig.EtchEndpoints) == 0 {
			explorer := defaultexplore.NewDefaultExplore()
			watcher, _ := explorer.GetDefaultWatcher(common.NodeExplorePath)
			watcher.SetPairs(utils.MapToSlice(ctx.Peers, func(id uint64, url string) serviceexplore.ServiceWatchKeyValPair {
				return serviceexplore.ServiceWatchKeyValPair{
					Key: common.GetPath(id),
					Val: url,
				}
			})...)
			return explorer
		}

		return etcdexplore.NewEtcdExplore(&etcdexplore.EtcdExploreConf{
			Context:     context.Background(),
			Endpoints:   ctx.AppConfig.EtchEndpoints,
			DialTimeout: 5 * time.Second,
		})
	})
	inject(c, NewServerExplorer)

	//raft
	inject(c, func(explore *ServerExplorer, engine *ServerEngine, c *common.AppContext) *raft.RaftServer {
		return raft.NewRaftServer(&raft.RaftServerConf{
			NodeID:        ctx.NodeID,
			URL:           ctx.NodeUrl,
			Peers:         ctx.Peers,
			HomePath:      path.Join(ctx.Home, ctx.DataDir),
			Ticker:        time.NewTicker(100 * time.Millisecond),
			ElectionTick:  10,
			HeartbeatTick: 1,
			SnapshotTick:  100,
			Engine:        engine,
			Logger:        log.NewZeroLogger(log.Debug),
			Explorer:      explore,
			Context:       context.Background(),
		})
	})
	inject(c, NewServerEngine)
	inject(c, helloworld.NewHelloWorldService)

	return c
}

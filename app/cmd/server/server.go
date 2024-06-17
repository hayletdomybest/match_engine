package server

import (
	"context"
	"fmt"
	"match_engine/app/cmd/common"
	"match_engine/app/helloworld"
	"match_engine/infra/consensus/raft"
	"match_engine/infra/db"
	"match_engine/infra/log"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type Server struct {
	Port      uint64
	engine    *gin.Engine
	container *dig.Container

	services []interface{}
}

func NewServer(ctx *common.AppContext) *Server {
	container := dig.New()

	r := gin.Default()
	group := r.Group("/api/v1")
	srv := &Server{
		Port:      ctx.ApiPort,
		engine:    r,
		container: container,
	}
	srv.inject(func() gin.IRoutes {
		return group
	})
	srv.inject(func() *common.AppContext {
		return ctx
	})

	//repository
	srv.inject(db.NewHelloWorldKv)
	srv.inject(db.NewInMemoryDBContext)

	//controller
	srv.inject(helloworld.NewHelloWorldController)

	//server
	srv.inject(func(engine *ServerEngine, c *common.AppContext) *raft.RaftServer {
		return raft.NewRaftServer(&raft.RaftServerConf{
			NodeID:        ctx.NodeID,
			URL:           ctx.URL,
			Peers:         ctx.Peers,
			HomePath:      path.Join(ctx.Home, ctx.DataDir),
			Ticker:        time.NewTicker(100 * time.Millisecond),
			ElectionTick:  10,
			HeartbeatTick: 1,
			SnapshotTick:  100,
			Engine:        engine,
			Logger:        log.NewZeroLogger(log.Debug),
			Context:       context.Background(),
		})
	})
	srv.inject(NewServerEngine)
	srv.inject(helloworld.NewHelloWorldService)

	return srv
}

func (server *Server) Run() error {
	for _, srv := range server.services {
		err := server.container.Invoke(srv)
		if err != nil {
			return err
		}
	}

	var raftServer *raft.RaftServer
	server.container.Invoke(func(engine *ServerEngine, srv *raft.RaftServer) {
		raftServer = srv
	})

	if err := raftServer.Start(); err != nil {
		return err
	}

	return server.engine.Run(fmt.Sprintf(":%d", server.Port))
}

func (server *Server) inject(srv interface{}) {
	server.services = append(server.services, srv)
	container := server.container
	container.Provide(srv)
}

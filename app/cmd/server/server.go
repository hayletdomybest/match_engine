package server

import (
	"errors"
	"fmt"
	"match_engine/app/cmd/common"
	"match_engine/infra/consensus/raft"
	serviceexplore "match_engine/infra/service_explore"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type ServerConf struct {
	dig.In

	Ctx        *common.AppContext
	RaftServer *raft.RaftServer
	Engine     *gin.Engine
	Explorer   serviceexplore.ServiceExplore
}

type Server struct {
	Port       uint64
	raftServer *raft.RaftServer
	engine     *gin.Engine
	explorer   serviceexplore.ServiceExplore
	serverUrl  string
	nodeID     uint64
	nodeURL    string
}

func NewServer(conf ServerConf) *Server {
	ctx := conf.Ctx
	srv := &Server{
		engine:     conf.Engine,
		explorer:   conf.Explorer,
		raftServer: conf.RaftServer,
		nodeID:     ctx.NodeID,
		nodeURL:    ctx.NodeUrl,
		serverUrl:  fmt.Sprintf("http://127.0.0.1:%d", ctx.ApiPort),
		Port:       ctx.ApiPort,
	}

	return srv
}

func (server *Server) Run() error {

	doneC := make(chan struct{})

	errC := make(chan error)

	go func() {
		err := server.engine.Run(fmt.Sprintf(":%d", server.Port))
		errC <- err
		close(doneC)
	}()

	go server.runRaftServer(errC, doneC)

	fmt.Println("server is running")

	running := true
	for running {
		select {
		case <-doneC:
			running = false
		case err := <-errC:
			fmt.Printf("server got error (%v)\n", err)
		}
	}
	fmt.Println("server has shutdown")
	return nil
}

func (server *Server) runRaftServer(errC chan error, doneC chan struct{}) {
	var raftServer *raft.RaftServer = server.raftServer

	if raftServer == nil {
		errC <- errors.New("raft server can not be nil")
		close(doneC)
		return
	}

	if err := raftServer.Start(); err != nil {
		errC <- fmt.Errorf("run start server error: %v", err)
		close(doneC)
		return
	}

	defer raftServer.Stop()
	<-doneC
}

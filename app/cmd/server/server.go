package server

import (
	"match_engine/app/helloworld"
	"match_engine/infra/db"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type Server struct {
	URL       string
	engine    *gin.Engine
	container *dig.Container

	services []interface{}
}

func NewServer(url string) *Server {
	r := gin.Default()
	group := r.Group("/api/v1")
	container := dig.New()

	container.Provide(func() gin.IRoutes {
		return group
	})

	return &Server{
		URL:       url,
		engine:    r,
		container: container,
	}
}

func (server *Server) Run() error {
	for _, srv := range server.services {
		server.container.Invoke(srv)
	}

	return server.engine.Run(server.URL)
}

func (server *Server) inject(srv interface{}) {
	server.services = append(server.services, srv)
	container := server.container
	container.Provide(srv)
}

func (server *Server) RegisterController() {
	server.inject(helloworld.NewHelloWorldController)
}

func (server *Server) RegisterRepository() {
	server.inject(db.NewHelloWorldKv)
	server.inject(db.NewInMemoryDBContext)
}

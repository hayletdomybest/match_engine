package helloworld

import (
	"match_engine/infra/db"
)

type HelloWorldService struct {
	dbContext *db.InMemoryDBContext
}

func NewHelloWorldService(dbContext *db.InMemoryDBContext) *HelloWorldService {
	return &HelloWorldService{
		dbContext: dbContext,
	}
}

func (engine *HelloWorldService) AppendMessage(msg string) error {
	engine.dbContext.HelloWorldKV.Append(msg)
	return nil
}

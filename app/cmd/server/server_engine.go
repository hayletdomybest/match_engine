package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"match_engine/app/cmd/common/model"
	"match_engine/app/helloworld"
	"match_engine/infra/consensus"
	"match_engine/infra/db"
	"match_engine/utils"
	"reflect"
	"sync"
	"time"
)

type ServerEngineHandler interface {
	Handle(any) error
}

type ServerEngine struct {
	actions       map[string]func(...any) error
	readActionMu  sync.Mutex
	readActions   map[string]chan<- uint64
	dbContext     *db.InMemoryDBContext
	helloworldSrv *helloworld.HelloWorldService
	idGenerator   utils.IDGenerator
	errorC        chan<- error
}

var _ consensus.CoordEngine = (*ServerEngine)(nil)

func NewServerEngine(
	dbContext *db.InMemoryDBContext,
	helloworldSrv *helloworld.HelloWorldService,
) *ServerEngine {
	engine := &ServerEngine{
		dbContext:     dbContext,
		helloworldSrv: helloworldSrv,
		actions:       make(map[string]func(...any) error),
		readActions:   make(map[string]chan<- uint64),
		idGenerator:   utils.NewSnowFlake(),
	}

	engine.actions[helloworld.ActionAppendMessage] = wrap(helloworldSrv.AppendMessage)

	return engine
}

func wrap(fn interface{}) func(...any) error {
	return func(args ...any) error {
		v := reflect.ValueOf(fn)
		t := v.Type()

		if len(args) != t.NumIn() {
			return errors.New("invalid number of arguments")
		}

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			if reflect.TypeOf(arg) != t.In(i) {
				return fmt.Errorf("argument %d must be of type %s", i, t.In(i))
			}
			in[i] = reflect.ValueOf(arg)
		}

		out := v.Call(in)

		if len(out) != 1 {
			return errors.New("function must return one value of type error")
		}
		if err, ok := out[0].Interface().(error); ok {
			return err
		}
		return errors.New("function return value is not of type error")
	}
}

func (engine *ServerEngine) Handle(data []byte) error {
	var msg model.AppMessage[any]
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	return engine.actions[msg.Action](msg.Data)
}

func (engine *ServerEngine) GenerateID() (uint64, error) {
	return engine.idGenerator.Generate()
}

func (engine *ServerEngine) GetSnapshot() ([]byte, error) {
	return engine.dbContext.CreateSnap()
}

func (engine *ServerEngine) ReloadSnapshot(bz []byte) error {
	return engine.dbContext.LoadSnap(bz)
}

func (engine *ServerEngine) SetErrorChan(ch chan<- error) {
	engine.errorC = ch
}

func (engine *ServerEngine) ReadHandle(index uint64, requestCtx []byte) error {
	rID := string(requestCtx)
	engine.readActionMu.Lock()
	readChan, existed := engine.readActions[rID]
	delete(engine.readActions, rID)
	engine.readActionMu.Unlock()
	if !existed {
		return fmt.Errorf("read request ID:%s not found", rID)
	}

	go func() {
		select {
		case readChan <- index:
		case <-time.After(100 * time.Second):
			if engine.errorC != nil {
				engine.errorC <- fmt.Errorf("read request ID:%s read time out", rID)
			}
		}
	}()
	return nil
}

func (engine *ServerEngine) CreateSyncRead(requestCtx []byte) chan uint64 {
	engine.readActionMu.Lock()
	c := make(chan uint64)
	engine.readActions[string(requestCtx)] = c
	engine.readActionMu.Unlock()
	return c
}

func (engine *ServerEngine) CancelSyncRead(requestCtx []byte) {
	rID := string(requestCtx)
	engine.readActionMu.Lock()
	delete(engine.readActions, rID)
	engine.readActionMu.Unlock()
}

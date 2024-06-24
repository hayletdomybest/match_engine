package helloworld

import (
	"encoding/json"
	"io"
	"match_engine/app/cmd/common"
	"match_engine/app/cmd/common/model"
	"match_engine/infra/consensus/raft"
	"match_engine/infra/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HelloWorldBody struct {
	Message string
}

type HelloWorldController struct {
	dbContext  *db.InMemoryDBContext // for test
	raftServer *raft.RaftServer
	appCtx     *common.AppContext
}

func NewHelloWorldController(
	r gin.IRoutes,
	context *common.AppContext,
	raftServer *raft.RaftServer,
	dbContext *db.InMemoryDBContext,
) *HelloWorldController {
	ctr := &HelloWorldController{
		dbContext:  dbContext,
		raftServer: raftServer,
		appCtx:     context,
	}
	r.POST("/helloworld/message", ctr.appendMessage)
	r.GET("/helloworld/messages", ctr.getMessages)
	r.GET("/helloworld/sync-messages", ctr.syncGetMessages)

	return ctr
}

func (ctr *HelloWorldController) appendMessage(c *gin.Context) {
	bz, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}
	var body HelloWorldBody
	if err := json.Unmarshal(bz, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to unmarshal body"})
		return
	}

	bz, _ = json.Marshal(&model.AppMessage[string]{
		Action: ActionAppendMessage,
		Data:   body.Message,
	})

	ctr.raftServer.Propose(bz)
	c.JSON(http.StatusOK, gin.H{"message": "Message appended successfully"})
}

func (ctr *HelloWorldController) getMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"messages": ctr.dbContext.HelloWorldKV.GetAll()})
}

func (ctr *HelloWorldController) syncGetMessages(c *gin.Context) {
	ch, err := ctr.raftServer.ReadIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"messages": "server read index error"})
	}

	select {
	case <-ch:
		c.JSON(http.StatusOK, gin.H{"messages": ctr.dbContext.HelloWorldKV.GetAll()})
	case <-time.After(2 * time.Second):
		c.JSON(http.StatusInternalServerError, gin.H{"messages": "server read index timeout"})
	}

}

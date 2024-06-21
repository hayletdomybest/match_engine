package explorer

import (
	"match_engine/app/cmd/common"
	"match_engine/infra/consensus/raft"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExplorerController struct {
	raftServer *raft.RaftServer
	appCtx     *common.AppContext
}

func NewExplorerController(
	r gin.IRoutes,
	context *common.AppContext,
	raftServer *raft.RaftServer,
) *ExplorerController {
	ctr := &ExplorerController{
		raftServer: raftServer,
		appCtx:     context,
	}
	r.GET("/explorer/leader", ctr.getLeader)
	return ctr
}

func (ctr *ExplorerController) getLeader(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"leader": ctr.raftServer.GetLeader()})
}

package helloworld

import (
	"encoding/json"
	"io"
	"match_engine/app/cmd/common"
	"match_engine/infra/db"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type HelloWorldBody struct {
	Message string
}

type HelloWorldController struct {
	dbCtx  *db.InMemoryDBContext
	appCtx *common.GlobalContext
}

func NewHelloWorldController(
	r gin.IRoutes,
	dbCtx *db.InMemoryDBContext,
	appCtx *common.GlobalContext,
) *HelloWorldController {
	ctr := &HelloWorldController{
		dbCtx:  dbCtx,
		appCtx: appCtx,
	}

	r.POST("/helloworld/message", ctr.appendMessage)
	r.GET("/helloworld/messages", ctr.getMessages)
	r.POST("/helloworld/create-snapshot", ctr.testCreateSnapshot)
	r.POST("/helloworld/load-snapshot", ctr.testLoadSnapshot)

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

	repo := ctr.dbCtx.HelloWorldKV

	repo.Append(body.Message)

	c.JSON(http.StatusOK, gin.H{"message": "Message appended successfully"})
}

func (ctr *HelloWorldController) getMessages(c *gin.Context) {
	repo := ctr.dbCtx.HelloWorldKV
	c.JSON(http.StatusOK, gin.H{"messages": repo.GetAll()})
}

var testsnapshot_file_name = "snapshot.json"

func (ctr *HelloWorldController) testCreateSnapshot(c *gin.Context) {
	bz, err := ctr.dbCtx.CreateSnap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	file := filepath.Join(ctr.appCtx.Home, testsnapshot_file_name)
	err = os.WriteFile(file, bz, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	c.String(http.StatusOK, "ok")
}

func (ctr *HelloWorldController) testLoadSnapshot(c *gin.Context) {
	fileName := filepath.Join(ctr.appCtx.Home, testsnapshot_file_name)

	// Read the config file
	file, err := os.Open(fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	defer file.Close()

	bz, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	ctr.dbCtx.LoadSnap(bz)
	repo := ctr.dbCtx.HelloWorldKV

	c.JSON(http.StatusOK, gin.H{"messages": repo.GetAll()})
}

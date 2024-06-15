package helloworld

import (
	"encoding/json"
	"io"
	"match_engine/infra/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HelloWorldBody struct {
	Message string
}

type HelloWorldController struct {
	kv *db.HelloWorldKv
}

func NewHelloWorldController(r gin.IRoutes, kv *db.HelloWorldKv) *HelloWorldController {
	ctr := &HelloWorldController{
		kv: kv,
	}

	r.POST("/helloworld/message", ctr.appendMessage)
	r.GET("/helloworld/messages", ctr.getMessages)

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

	ctr.kv.Append(body.Message)

	c.JSON(http.StatusOK, gin.H{"message": "Message appended successfully"})
}

func (ctr *HelloWorldController) getMessages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"messages": ctr.kv.GetAll()})
}

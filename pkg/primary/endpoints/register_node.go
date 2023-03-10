package endpoints

import (
	"fmt"

	node2 "keepair/pkg/primary/node"

	"github.com/gin-gonic/gin"
)

type RegisterNodeBody struct {
	ID   string `json:"id" binding:"required"`
	Port string `json:"port" binding:"required"`
}

var RegisterNodeHandler = func(nodeService node2.IService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body RegisterNodeBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.Data(400, "", []byte(fmt.Sprintf("error: %s", err.Error())))
			return
		}

		err := nodeService.RegisterNode(node2.NewNode(body.ID, c.ClientIP(), body.Port))
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

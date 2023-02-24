package endpoints

import (
	"fmt"

	"keepair/pkg/node"

	"github.com/gin-gonic/gin"
)

type RegisterNodeBody struct {
	ID   string `json:"id" binding:"required"`
	Port string `json:"port" binding:"required"`
}

var RegisterNodeHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body RegisterNodeBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.Data(400, "", []byte(fmt.Sprintf("error: %s", err.Error())))
			return
		}
		nodeService.RegisterNode(node.NewNode(body.ID, c.ClientIP(), body.Port))
	}
}

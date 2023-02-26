package endpoints

import (
	"keepair/pkg/node"

	"github.com/gin-gonic/gin"
)

var GetNodesHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nodes := nodeService.GetNodes()
		c.JSON(200, gin.H{
			"nodes": nodes,
		})
	}
}

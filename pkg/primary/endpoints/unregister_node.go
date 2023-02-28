package endpoints

import (
	"keepair/pkg/primary/node"

	"github.com/gin-gonic/gin"
)

var UnregisterNodeHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {

		nodeID := c.Param("nodeID")
		if nodeID == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		numNodes := nodeService.GetNumNodes()
		if numNodes == 0 {
			c.Data(500, "", []byte("no nodes available"))
			return
		}

		err := nodeService.UnregisterNode(nodeID)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

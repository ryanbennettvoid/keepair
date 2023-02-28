package endpoints

import (
	"fmt"

	"keepair/pkg/partition"
	"keepair/pkg/primary/clients"
	"keepair/pkg/primary/node"

	"github.com/gin-gonic/gin"
)

var GetKeyHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		numNodes := nodeService.GetNumNodes()
		if numNodes == 0 {
			c.Data(500, "", []byte("no nodes available"))
			return
		}

		partitionKey := partition.GenerateDeterministicPartitionKey(key, numNodes)
		n, err := nodeService.GetNodeByIndex(partitionKey)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		workerNodeURL := fmt.Sprintf("http://%s", n.Address)
		workerClient := clients.NewWorkerClient(workerNodeURL)
		value, err := workerClient.GetKey(key)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", value)
	}
}

package endpoints

import (
	"fmt"

	"keepair/pkg/clients"
	"keepair/pkg/node"
	"keepair/pkg/partition"

	"github.com/gin-gonic/gin"
)

var DeleteKeyHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		numNodes := nodeService.GetNumNodes()
		partitionKey := partition.GenerateDeterministicPartitionKey(key, numNodes)
		n, err := nodeService.GetNodeByIndex(partitionKey)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		workerNodeURL := fmt.Sprintf("http://%s", n.Address)
		workerClient := clients.NewWorkerClient(workerNodeURL)
		if err := workerClient.DeleteKey(key); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

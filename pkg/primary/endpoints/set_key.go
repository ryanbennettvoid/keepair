package endpoints

import (
	"fmt"
	"io"

	"keepair/pkg/clients"
	"keepair/pkg/node"
	"keepair/pkg/partition"

	"github.com/gin-gonic/gin"
)

var SetKeyHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		numNodes := nodeService.GetNumNodes()
		partitionKey := partition.GetDeterministicPartitionKey(key, numNodes)
		n, err := nodeService.GetNodeByIndex(partitionKey)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		postBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}
		defer c.Request.Body.Close()

		workerNodeURL := fmt.Sprintf("http://%s", n.Address)
		workerClient := clients.NewWorkerClient(workerNodeURL)
		if err := workerClient.SetKey(key, postBody); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}
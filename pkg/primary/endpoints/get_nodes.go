package endpoints

import (
	"sync"

	"keepair/pkg/log"
	"keepair/pkg/node"

	"github.com/gin-gonic/gin"
)

var GetNodesHandler = func(nodeService node.IService) gin.HandlerFunc {
	return func(c *gin.Context) {

		mu := sync.Mutex{}
		nodes := nodeService.GetNodes()

		wg := sync.WaitGroup{}
		for i, n := range nodes {
			wg.Add(1)
			go func(i int, n node.Node) {
				defer wg.Done()

				err := n.LoadStats()
				if err != nil {
					log.Get().Printf("LoadStats ERR: %s", err)
				}

				mu.Lock()
				nodes[i] = n
				mu.Unlock()
			}(i, n)
		}
		wg.Wait()

		log.Get().Printf("NODES: %+v", nodes)

		c.JSON(200, gin.H{
			"nodes": nodes,
		})
	}
}

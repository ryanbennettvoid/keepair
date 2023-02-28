package endpoints

import (
	"keepair/pkg/common"
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

var GetStatsHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		stats := common.NodeStats{
			ObjectCount: store.GetObjectCount(),
		}

		c.JSON(200, gin.H{
			"stats": stats,
		})
	}
}

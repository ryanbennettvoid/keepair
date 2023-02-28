package endpoints

import (
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

var ApplyOperationsHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := store.ApplyOperations(); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

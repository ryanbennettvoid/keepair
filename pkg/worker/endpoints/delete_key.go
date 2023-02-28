package endpoints

import (
	"keepair/pkg/log"
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

var DeleteKeyHandler = func(workerID string, store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		log.Get().Printf("DELETE: %s on %s", key, workerID)
		if err := store.Delete(key); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}
		log.Get().Printf("DELETE DONE: %s on %s", key, workerID)

		c.Data(200, "", []byte("ok"))
	}
}

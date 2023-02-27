package endpoints

import (
	"keepair/pkg/store"

	"github.com/gin-gonic/gin"
)

var DeleteKeyHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		if err := store.Delete(key); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

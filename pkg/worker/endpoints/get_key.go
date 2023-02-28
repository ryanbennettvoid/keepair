package endpoints

import (
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

var GetKeyHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		value, err := store.Get(key)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", value)
	}
}

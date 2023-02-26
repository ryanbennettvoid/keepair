package endpoints

import (
	"io"

	"keepair/pkg/store"

	"github.com/gin-gonic/gin"
)

var SetKeyHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		key := c.Param("key")
		if key == "" {
			c.Data(400, "", []byte("empty key"))
			return
		}

		value, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}
		defer c.Request.Body.Close()

		if len(value) == 0 {
			c.Data(400, "", []byte("empty value"))
			return
		}

		if err := store.Set(key, value); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

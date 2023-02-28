package endpoints

import (
	"encoding/json"
	"io"

	"keepair/pkg/common"
	"keepair/pkg/worker/store"

	"github.com/gin-gonic/gin"
)

var QueueOperationsHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}
		defer c.Request.Body.Close()

		var operations []common.EntryOperation
		if err := json.Unmarshal(body, &operations); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		if err := store.QueueOperations(operations); err != nil {
			c.Data(500, "", []byte(err.Error()))
			return
		}

		c.Data(200, "", []byte("ok"))
	}
}

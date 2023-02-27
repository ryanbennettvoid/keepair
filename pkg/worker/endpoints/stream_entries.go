package endpoints

import (
	"fmt"

	"keepair/pkg/store"
	"keepair/pkg/streamer"

	"github.com/gin-gonic/gin"
)

var StreamEntriesHandler = func(store store.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.WriteHeaderNow()
		c.Writer.Flush()

		entryChan := store.StreamEntries()
		for entry := range entryChan {
			encodedMessage, err := streamer.EncodeMessage(entry)
			if err != nil {
				panic(fmt.Errorf("error encoding message: %w", err))
			}
			if _, err := c.Writer.WriteString(encodedMessage); err != nil {
				panic(fmt.Errorf("error writing string: %w", err))
			}
			c.Writer.Flush()
		}

	}
}

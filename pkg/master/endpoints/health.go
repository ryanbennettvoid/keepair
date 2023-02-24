package endpoints

import "github.com/gin-gonic/gin"

var HealthHandler = func() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(200, "", []byte("ok"))
	}
}

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter wires up overseer HTTP routes.
//
// API surface:
//
//	GET /health — liveness probe
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

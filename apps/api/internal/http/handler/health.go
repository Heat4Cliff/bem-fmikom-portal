package handler

import "github.com/gin-gonic/gin"

// Healthz is an operational liveness endpoint, not a product API feature.
func Healthz(c *gin.Context) {
	c.JSON(200, Success("Service is healthy", gin.H{"status": "ok"}))
}

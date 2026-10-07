package router

import (
	"log/slog"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/handler"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

func New(logger *slog.Logger) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.RequestLogger(logger))
	engine.GET("/healthz", handler.Healthz)
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, handler.Failure("Resource not found", nil))
	})
	return engine
}

package middleware

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request", "request_id", c.GetHeader("X-Request-ID"), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "latency_ms", time.Since(start).Milliseconds())
	}
}

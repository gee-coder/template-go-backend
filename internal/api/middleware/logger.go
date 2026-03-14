package middleware

import (
	"time"

	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLog logs the HTTP request metadata.
func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http access",
			zap.String("requestID", utils.RequestIDFromContext(c)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("clientIP", c.ClientIP()),
		)
	}
}


package middleware

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery recovers from panics and logs them.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("panic recovered", zap.Any("panic", recovered))
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Envelope{
			Code:      http.StatusInternalServerError,
			Message:   "服务器繁忙，请稍后重试",
			RequestID: utils.RequestIDFromContext(c),
		})
	})
}


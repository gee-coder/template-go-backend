package handler

import (
	"time"

	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check APIs.
type HealthHandler struct{}

// NewHealthHandler creates a health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check reports the service health status.
func (h *HealthHandler) Check(c *gin.Context) {
	utils.RespondOK(c, gin.H{
		"status": "正常",
		"time":   time.Now(),
	})
}

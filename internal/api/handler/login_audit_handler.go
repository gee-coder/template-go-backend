package handler

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// LoginAuditHandler handles login audit APIs.
type LoginAuditHandler struct {
	service LoginAuditService
}

// NewLoginAuditHandler creates a login audit handler.
func NewLoginAuditHandler(service LoginAuditService) *LoginAuditHandler {
	return &LoginAuditHandler{service: service}
}

// List returns login audit logs.
func (h *LoginAuditHandler) List(c *gin.Context) {
	var query request.LoginAuditListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	items, err := h.service.List(c.Request.Context(), repository.LoginAuditFilter{
		Keyword:   query.Keyword,
		Status:    query.Status,
		LoginType: query.LoginType,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondOK(c, items)
}

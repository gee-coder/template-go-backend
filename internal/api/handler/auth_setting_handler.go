package handler

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/service"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthSettingHandler handles admin auth setting APIs.
type AuthSettingHandler struct {
	service AuthSettingService
}

// NewAuthSettingHandler creates an auth setting handler.
func NewAuthSettingHandler(service AuthSettingService) *AuthSettingHandler {
	return &AuthSettingHandler{service: service}
}

// Get returns current auth settings.
func (h *AuthSettingHandler) Get(c *gin.Context) {
	settings, err := h.service.Get(c.Request.Context())
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, settings)
}

// Update updates auth settings.
func (h *AuthSettingHandler) Update(c *gin.Context) {
	var req request.UpdateAuthSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	settings, err := h.service.Update(c.Request.Context(), service.UpdateAuthSettingInput{
		EnableEmailLogin:        req.EnableEmailLogin,
		EnablePhoneLogin:        req.EnablePhoneLogin,
		EnableEmailRegistration: req.EnableEmailRegistration,
		EnablePhoneRegistration: req.EnablePhoneRegistration,
		EnableTwoFactor:         req.EnableTwoFactor,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, settings)
}

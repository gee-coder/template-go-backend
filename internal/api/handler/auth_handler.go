package handler

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/middleware"
	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles auth APIs.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// ResolvePermissions resolves the current user permissions.
func (h *AuthHandler) ResolvePermissions(c *gin.Context, userID uint) ([]string, error) {
	return h.authService.ResolvePermissions(c.Request.Context(), userID)
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// Refresh handles access token refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.RefreshTokenRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, gin.H{"success": true})
}

// Profile returns the current user profile.
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := middleware.MustUserID(c)
	profile, err := h.authService.Profile(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, profile)
}

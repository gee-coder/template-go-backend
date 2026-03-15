package handler

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/service"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// BrandingSettingHandler handles brand setting APIs.
type BrandingSettingHandler struct {
	service BrandingSettingService
}

// NewBrandingSettingHandler creates a branding setting handler.
func NewBrandingSettingHandler(service BrandingSettingService) *BrandingSettingHandler {
	return &BrandingSettingHandler{service: service}
}

// GetPublic returns public brand settings for login and other public pages.
func (h *BrandingSettingHandler) GetPublic(c *gin.Context) {
	settings, err := h.service.Get(c.Request.Context())
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, settings)
}

// Get returns current admin brand settings.
func (h *BrandingSettingHandler) Get(c *gin.Context) {
	settings, err := h.service.Get(c.Request.Context())
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, settings)
}

// Update updates current admin brand settings.
func (h *BrandingSettingHandler) Update(c *gin.Context) {
	var req request.UpdateBrandingSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	settings, err := h.service.Update(c.Request.Context(), service.UpdateBrandingSettingInput{
		AppTitle:       req.AppTitle,
		ConsoleName:    req.ConsoleName,
		ProductTagline: req.ProductTagline,
		LogoMarkURL:    req.LogoMarkURL,
		FaviconURL:     req.FaviconURL,
		LoginHeroURL:   req.LoginHeroURL,
		Theme: service.BrandingTheme{
			Primary:     req.Theme.Primary,
			PrimaryDark: req.Theme.PrimaryDark,
			ShellStart:  req.Theme.ShellStart,
			ShellEnd:    req.Theme.ShellEnd,
			HeroStart:   req.Theme.HeroStart,
			HeroEnd:     req.Theme.HeroEnd,
		},
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, settings)
}

// UploadAsset uploads a brand asset and returns its URL.
func (h *BrandingSettingHandler) UploadAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "请选择要上传的图片"))
		return
	}

	payload, err := h.service.UploadAsset(c.Request.Context(), service.UploadBrandingAssetInput{
		Kind: c.PostForm("kind"),
		File: file,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

package handler

import (
	"net/http"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/api/middleware"
	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/service"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles auth APIs.
type AuthHandler struct {
	authService         AuthService
	loginAuditService   LoginAuditService
	avatarAssetService  AvatarAssetService
	smsService          SMSVerificationService
	emailService        EmailVerificationService
	imageCaptchaService ImageCaptchaService
}

// NewAuthHandler creates an auth handler.
func NewAuthHandler(
	authService AuthService,
	loginAuditService LoginAuditService,
	avatarAssetService AvatarAssetService,
	smsService SMSVerificationService,
	emailService EmailVerificationService,
	imageCaptchaService ImageCaptchaService,
) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		loginAuditService:   loginAuditService,
		avatarAssetService:  avatarAssetService,
		smsService:          smsService,
		emailService:        emailService,
		imageCaptchaService: imageCaptchaService,
	}
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

	account := req.Account
	if account == "" {
		account = req.Username
	}
	account = strings.TrimSpace(account)
	if account == "" {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "account is required"))
		return
	}

	payload, err := h.authService.Login(c.Request.Context(), service.LoginInput{
		Account:          account,
		LoginType:        req.LoginType,
		Password:         req.Password,
		VerificationCode: req.VerificationCode,
		CaptchaID:        req.CaptchaID,
		CaptchaCode:      req.CaptchaCode,
		TwoFactorCode:    req.TwoFactorCode,
	})
	h.writeLoginAudit(c, account, req.LoginType, payload, err)
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// Register handles public user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Account:          req.Account,
		RegisterType:     req.RegisterType,
		Nickname:         req.Nickname,
		Password:         req.Password,
		VerificationCode: req.VerificationCode,
		CaptchaID:        req.CaptchaID,
		CaptchaCode:      req.CaptchaCode,
		SMSCode:          req.SMSCode,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondCreated(c, payload)
}

// Captcha returns a generated image captcha.
func (h *AuthHandler) Captcha(c *gin.Context) {
	payload, err := h.imageCaptchaService.Create(c.Request.Context())
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// SendSMSCode sends a phone verification code through the configured provider.
func (h *AuthHandler) SendSMSCode(c *gin.Context) {
	var req request.SendSMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.smsService.SendCode(c.Request.Context(), service.SendSMSCodeInput{
		Phone:   req.Phone,
		Purpose: req.Purpose,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// SendEmailCode sends an email verification code through the configured provider.
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req request.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.emailService.SendCode(c.Request.Context(), service.SendEmailCodeInput{
		Email:   req.Email,
		Purpose: req.Purpose,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// SendTwoFactorCode sends a second-factor verification code for username login.
func (h *AuthHandler) SendTwoFactorCode(c *gin.Context) {
	var req request.SendTwoFactorCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	payload, err := h.authService.SendTwoFactorCode(c.Request.Context(), service.SendTwoFactorCodeInput{
		Account:   req.Account,
		LoginType: req.LoginType,
	})
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

// UpdateProfile updates the current user profile.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	profile, err := h.authService.UpdateProfile(c.Request.Context(), userID, service.UpdateProfileInput{
		Avatar: req.Avatar,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, profile)
}

// UploadAvatarAsset uploads a custom avatar image for authenticated users.
func (h *AuthHandler) UploadAvatarAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "please choose an avatar image"))
		return
	}

	payload, err := h.avatarAssetService.Upload(c.Request.Context(), service.UploadAvatarAssetInput{File: file})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, payload)
}

// Options returns public auth options for clients.
func (h *AuthHandler) Options(c *gin.Context) {
	options, err := h.authService.Options(c.Request.Context())
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, options)
}

func (h *AuthHandler) writeLoginAudit(c *gin.Context, account string, requestedType string, payload *service.TokenPayload, loginErr error) {
	if h.loginAuditService == nil {
		return
	}

	input := service.CreateLoginAuditInput{
		Account:   account,
		LoginType: inferLoginAuditType(requestedType, account),
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Status:    "success",
		Message:   "login success",
	}

	if payload != nil && payload.User != nil {
		input.UserID = &payload.User.ID
		input.Username = payload.User.Username
	}
	if loginErr != nil {
		input.Status = "failure"
		input.Message = loginErr.Error()
	}

	_ = h.loginAuditService.Create(c.Request.Context(), input)
}

func inferLoginAuditType(requestedType string, account string) string {
	requestedType = strings.TrimSpace(requestedType)
	switch requestedType {
	case "email", "phone", "username":
		return requestedType
	}

	account = strings.TrimSpace(account)
	if account == "" {
		return "username"
	}
	if strings.Contains(account, "@") {
		return "email"
	}

	normalized := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "").Replace(account)
	isPhone := true
	for _, char := range normalized {
		if char < '0' || char > '9' {
			isPhone = false
			break
		}
	}
	if isPhone && len(normalized) >= 6 {
		return "phone"
	}

	return "username"
}

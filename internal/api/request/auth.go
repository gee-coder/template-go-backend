package request

// LoginRequest describes the login payload.
type LoginRequest struct {
	Account          string `json:"account" binding:"omitempty,min=3,max=128"`
	Username         string `json:"username" binding:"omitempty,min=3,max=128"`
	LoginType        string `json:"loginType" binding:"omitempty,oneof=username email phone"`
	Password         string `json:"password" binding:"omitempty,min=6,max=64"`
	VerificationCode string `json:"verificationCode" binding:"omitempty,min=4,max=8"`
	CaptchaID        string `json:"captchaId" binding:"omitempty,min=8,max=128"`
	CaptchaCode      string `json:"captchaCode" binding:"omitempty,min=4,max=8"`
	TwoFactorCode    string `json:"twoFactorCode" binding:"omitempty,min=4,max=8"`
}

// RegisterRequest describes the register payload.
type RegisterRequest struct {
	Account          string `json:"account" binding:"required,min=3,max=128"`
	RegisterType     string `json:"registerType" binding:"omitempty,oneof=email phone"`
	Nickname         string `json:"nickname" binding:"omitempty,min=2,max=64"`
	Password         string `json:"password" binding:"required,min=6,max=64"`
	VerificationCode string `json:"verificationCode" binding:"omitempty,min=4,max=8"`
	CaptchaID        string `json:"captchaId" binding:"omitempty,min=8,max=128"`
	CaptchaCode      string `json:"captchaCode" binding:"omitempty,min=4,max=8"`
	SMSCode          string `json:"smsCode" binding:"omitempty,min=4,max=8"`
}

// SendSMSCodeRequest describes the SMS send-code payload.
type SendSMSCodeRequest struct {
	Phone   string `json:"phone" binding:"required,min=6,max=20"`
	Purpose string `json:"purpose" binding:"required,oneof=register login bind_phone reset_password two_factor"`
}

// SendEmailCodeRequest describes the email send-code payload.
type SendEmailCodeRequest struct {
	Email   string `json:"email" binding:"required,email,max=128"`
	Purpose string `json:"purpose" binding:"required,oneof=login register two_factor"`
}

// SendTwoFactorCodeRequest describes the two-factor send-code payload.
type SendTwoFactorCodeRequest struct {
	Account   string `json:"account" binding:"required,min=3,max=128"`
	LoginType string `json:"loginType" binding:"omitempty,oneof=username"`
}

// UpdateProfileRequest describes the self profile update payload.
type UpdateProfileRequest struct {
	Avatar string `json:"avatar" binding:"required,min=3,max=255"`
}

// RefreshTokenRequest describes the refresh token payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

package request

// LoginRequest describes the login payload.
type LoginRequest struct {
	Account   string `json:"account" binding:"omitempty,min=3,max=128"`
	Username  string `json:"username" binding:"omitempty,min=3,max=128"`
	LoginType string `json:"loginType" binding:"omitempty,oneof=username email phone"`
	Password  string `json:"password" binding:"required,min=6,max=64"`
	SMSCode   string `json:"smsCode" binding:"omitempty,min=4,max=8"`
}

// RegisterRequest describes the register payload.
type RegisterRequest struct {
	Account      string `json:"account" binding:"required,min=3,max=128"`
	RegisterType string `json:"registerType" binding:"omitempty,oneof=email phone"`
	Nickname     string `json:"nickname" binding:"omitempty,min=2,max=64"`
	Password     string `json:"password" binding:"required,min=6,max=64"`
	SMSCode      string `json:"smsCode" binding:"omitempty,min=4,max=8"`
}

// SendSMSCodeRequest describes the SMS send-code payload.
type SendSMSCodeRequest struct {
	Phone   string `json:"phone" binding:"required,min=6,max=20"`
	Purpose string `json:"purpose" binding:"required,oneof=register login bind_phone reset_password"`
}

// VerifySMSCodeRequest describes the SMS verify-code payload.
type VerifySMSCodeRequest struct {
	Phone   string `json:"phone" binding:"required,min=6,max=20"`
	Purpose string `json:"purpose" binding:"required,oneof=register login bind_phone reset_password"`
	Code    string `json:"code" binding:"required,min=4,max=8"`
}

// UpdateProfileRequest describes the self profile update payload.
type UpdateProfileRequest struct {
	Avatar string `json:"avatar" binding:"required,min=3,max=255"`
}

// RefreshTokenRequest describes the refresh token payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

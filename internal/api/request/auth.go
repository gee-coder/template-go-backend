package request

// LoginRequest describes the login payload.
type LoginRequest struct {
	Account   string `json:"account" binding:"omitempty,min=3,max=128"`
	Username  string `json:"username" binding:"omitempty,min=3,max=128"`
	LoginType string `json:"loginType" binding:"omitempty,oneof=username email phone"`
	Password  string `json:"password" binding:"required,min=6,max=64"`
}

// RegisterRequest describes the register payload.
type RegisterRequest struct {
	Account      string `json:"account" binding:"required,min=3,max=128"`
	RegisterType string `json:"registerType" binding:"required,oneof=email phone"`
	Nickname     string `json:"nickname" binding:"omitempty,min=2,max=64"`
	Password     string `json:"password" binding:"required,min=6,max=64"`
}

// RefreshTokenRequest describes the refresh token payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

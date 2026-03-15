package request

// UpdateAuthSettingRequest describes auth setting updates from admin.
type UpdateAuthSettingRequest struct {
	EnableEmailLogin        bool `json:"enableEmailLogin"`
	EnablePhoneLogin        bool `json:"enablePhoneLogin"`
	EnableEmailRegistration bool `json:"enableEmailRegistration"`
	EnablePhoneRegistration bool `json:"enablePhoneRegistration"`
}

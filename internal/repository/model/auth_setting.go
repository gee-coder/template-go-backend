package model

// AuthSetting stores runtime auth channel switches managed from admin.
type AuthSetting struct {
	BaseModel
	EnableEmailLogin        bool `gorm:"default:true" json:"enableEmailLogin"`
	EnablePhoneLogin        bool `gorm:"default:true" json:"enablePhoneLogin"`
	EnableEmailRegistration bool `gorm:"default:true" json:"enableEmailRegistration"`
	EnablePhoneRegistration bool `gorm:"default:true" json:"enablePhoneRegistration"`
}

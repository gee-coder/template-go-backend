package model

// LoginAuditLog describes a login audit event.
type LoginAuditLog struct {
	BaseModel
	UserID    *uint  `gorm:"index" json:"userId"`
	Username  string `gorm:"size:64;index" json:"username"`
	Account   string `gorm:"size:128;index;not null" json:"account"`
	LoginType string `gorm:"size:32;index;not null" json:"loginType"`
	Status    string `gorm:"size:32;index;not null" json:"status"`
	IP        string `gorm:"size:64" json:"ip"`
	UserAgent string `gorm:"size:255" json:"userAgent"`
	Message   string `gorm:"size:255" json:"message"`
}

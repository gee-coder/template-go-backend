package model

// User describes the system user.
type User struct {
	BaseModel
	Username string `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Nickname string `gorm:"size:64;not null" json:"nickname"`
	Email    string `gorm:"size:128;uniqueIndex" json:"email"`
	Phone    string `gorm:"size:32;uniqueIndex" json:"phone"`
	Avatar   string `gorm:"size:255;default:'default-01'" json:"avatar"`
	Status   string `gorm:"size:32;default:enabled" json:"status"`
	Password string `gorm:"size:255;not null" json:"-"`
	Roles    []Role `gorm:"many2many:user_roles;" json:"roles"`
}

// UserRole describes the user-role relationship.
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

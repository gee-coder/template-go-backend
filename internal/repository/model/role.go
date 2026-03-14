package model

// Role describes the RBAC role.
type Role struct {
	BaseModel
	Name   string `gorm:"size:64;not null" json:"name"`
	Code   string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Status string `gorm:"size:32;default:enabled" json:"status"`
	Remark string `gorm:"size:255" json:"remark"`
	Menus  []Menu `gorm:"many2many:role_menus;" json:"menus"`
}

// RoleMenu describes the role-menu relationship.
type RoleMenu struct {
	RoleID uint `gorm:"primaryKey"`
	MenuID uint `gorm:"primaryKey"`
}


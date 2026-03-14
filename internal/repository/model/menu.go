package model

// Menu describes a menu item or permission node.
type Menu struct {
	BaseModel
	ParentID   uint   `gorm:"index" json:"parentId"`
	Name       string `gorm:"size:64;not null" json:"name"`
	Title      string `gorm:"size:64;not null" json:"title"`
	Path       string `gorm:"size:255" json:"path"`
	Component  string `gorm:"size:255" json:"component"`
	Icon       string `gorm:"size:64" json:"icon"`
	Type       string `gorm:"size:32;default:menu" json:"type"`
	Permission string `gorm:"size:128" json:"permission"`
	Sort       int    `gorm:"default:0" json:"sort"`
	Hidden     bool   `gorm:"default:false" json:"hidden"`
	Status     string `gorm:"size:32;default:enabled" json:"status"`
}


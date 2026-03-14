package request

// MenuListQuery describes the menu list query.
type MenuListQuery struct {
	Keyword string `form:"keyword"`
}

// CreateMenuRequest describes the create menu payload.
type CreateMenuRequest struct {
	ParentID   uint   `json:"parentId"`
	Name       string `json:"name" binding:"required,min=2,max=64"`
	Title      string `json:"title" binding:"required,min=2,max=64"`
	Path       string `json:"path" binding:"omitempty,max=255"`
	Component  string `json:"component" binding:"omitempty,max=255"`
	Icon       string `json:"icon" binding:"omitempty,max=64"`
	Type       string `json:"type" binding:"omitempty,oneof=menu button directory"`
	Permission string `json:"permission" binding:"omitempty,max=128"`
	Sort       int    `json:"sort"`
	Hidden     bool   `json:"hidden"`
	Status     string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

// UpdateMenuRequest describes the update menu payload.
type UpdateMenuRequest = CreateMenuRequest


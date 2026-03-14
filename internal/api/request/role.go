package request

// RoleListQuery describes the role list query.
type RoleListQuery struct {
	Keyword string `form:"keyword"`
	Status  string `form:"status"`
}

// CreateRoleRequest describes the create role payload.
type CreateRoleRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=64"`
	Code    string `json:"code" binding:"required,min=2,max=64"`
	Status  string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Remark  string `json:"remark" binding:"omitempty,max=255"`
	MenuIDs []uint `json:"menuIds"`
}

// UpdateRoleRequest describes the update role payload.
type UpdateRoleRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=64"`
	Status  string `json:"status" binding:"required,oneof=enabled disabled"`
	Remark  string `json:"remark" binding:"omitempty,max=255"`
	MenuIDs []uint `json:"menuIds"`
}


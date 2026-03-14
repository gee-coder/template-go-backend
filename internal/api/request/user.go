package request

// UserListQuery describes the user list query.
type UserListQuery struct {
	Keyword string `form:"keyword"`
	Status  string `form:"status"`
}

// CreateUserRequest describes the create user payload.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Nickname string `json:"nickname" binding:"required,min=2,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=32"`
	Status   string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	RoleIDs  []uint `json:"roleIds"`
}

// UpdateUserRequest describes the update user payload.
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=64"`
	Email    string `json:"email" binding:"omitempty,email,max=128"`
	Phone    string `json:"phone" binding:"omitempty,max=32"`
	Status   string `json:"status" binding:"required,oneof=enabled disabled"`
	Password string `json:"password" binding:"omitempty,min=6,max=64"`
	RoleIDs  []uint `json:"roleIds"`
}


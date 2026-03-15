package handler

import (
	"net/http"
	"strconv"

	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/service"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user management APIs.
type UserHandler struct {
	service UserService
}

// NewUserHandler creates a user handler.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// List returns users.
func (h *UserHandler) List(c *gin.Context) {
	var query request.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	users, err := h.service.List(c.Request.Context(), repository.UserFilter{
		Keyword: query.Keyword,
		Status:  query.Status,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, users)
}

// Create creates a new user.
func (h *UserHandler) Create(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	user, err := h.service.Create(c.Request.Context(), service.CreateUserInput{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
		Password: req.Password,
		RoleIDs:  req.RoleIDs,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondCreated(c, user)
}

// Update updates a user.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "用户 ID 不合法"))
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	user, err := h.service.Update(c.Request.Context(), uint(id), service.UpdateUserInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
		Password: req.Password,
		RoleIDs:  req.RoleIDs,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, user)
}

// Delete deletes a user.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "用户 ID 不合法"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, gin.H{"success": true})
}

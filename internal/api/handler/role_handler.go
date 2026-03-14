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

// RoleHandler handles role management APIs.
type RoleHandler struct {
	service RoleService
}

// NewRoleHandler creates a role handler.
func NewRoleHandler(service RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// List returns roles.
func (h *RoleHandler) List(c *gin.Context) {
	var query request.RoleListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	roles, err := h.service.List(c.Request.Context(), repository.RoleFilter{
		Keyword: query.Keyword,
		Status:  query.Status,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, roles)
}

// Create creates a role.
func (h *RoleHandler) Create(c *gin.Context) {
	var req request.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	role, err := h.service.Create(c.Request.Context(), service.CreateRoleInput{
		Name:    req.Name,
		Code:    req.Code,
		Status:  req.Status,
		Remark:  req.Remark,
		MenuIDs: req.MenuIDs,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondCreated(c, role)
}

// Update updates a role.
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "角色 ID 不合法"))
		return
	}

	var req request.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	role, err := h.service.Update(c.Request.Context(), uint(id), service.UpdateRoleInput{
		Name:    req.Name,
		Status:  req.Status,
		Remark:  req.Remark,
		MenuIDs: req.MenuIDs,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, role)
}

// Delete deletes a role.
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "角色 ID 不合法"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, gin.H{"success": true})
}

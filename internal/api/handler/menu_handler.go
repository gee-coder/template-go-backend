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

// MenuHandler handles menu management APIs.
type MenuHandler struct {
	service MenuService
}

// NewMenuHandler creates a menu handler.
func NewMenuHandler(service MenuService) *MenuHandler {
	return &MenuHandler{service: service}
}

// List returns menus.
func (h *MenuHandler) List(c *gin.Context) {
	var query request.MenuListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	menus, err := h.service.List(c.Request.Context(), repository.MenuFilter{
		Keyword: query.Keyword,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, menus)
}

// Create creates a menu.
func (h *MenuHandler) Create(c *gin.Context) {
	var req request.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	menu, err := h.service.Create(c.Request.Context(), service.CreateMenuInput{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Title:      req.Title,
		Path:       req.Path,
		Component:  req.Component,
		Icon:       req.Icon,
		Type:       req.Type,
		Permission: req.Permission,
		Sort:       req.Sort,
		Hidden:     req.Hidden,
		Status:     req.Status,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondCreated(c, menu)
}

// Update updates a menu.
func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "菜单 ID 不合法"))
		return
	}

	var req request.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	menu, err := h.service.Update(c.Request.Context(), uint(id), service.UpdateMenuInput{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Title:      req.Title,
		Path:       req.Path,
		Component:  req.Component,
		Icon:       req.Icon,
		Type:       req.Type,
		Permission: req.Permission,
		Sort:       req.Sort,
		Hidden:     req.Hidden,
		Status:     req.Status,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, menu)
}

// Delete deletes a menu.
func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "菜单 ID 不合法"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondOK(c, gin.H{"success": true})
}

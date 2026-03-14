package handler

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/request"
	"github.com/gee-coder/template-go-backend/internal/service"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// ContactHandler handles website contact APIs.
type ContactHandler struct {
	service ContactService
}

// NewContactHandler creates a contact handler.
func NewContactHandler(service ContactService) *ContactHandler {
	return &ContactHandler{service: service}
}

// Create creates a contact submission.
func (h *ContactHandler) Create(c *gin.Context) {
	var req request.CreateContactSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, utils.BindErrorMessage(err)))
		return
	}

	submission, err := h.service.Create(c.Request.Context(), service.CreateContactSubmissionInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Company: req.Company,
		Message: req.Message,
		Source:  req.Source,
	})
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.RespondCreated(c, submission)
}

package request

// CreateContactSubmissionRequest describes the public contact form payload.
type CreateContactSubmissionRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=64"`
	Email   string `json:"email" binding:"required,email,max=128"`
	Phone   string `json:"phone" binding:"omitempty,max=32"`
	Company string `json:"company" binding:"omitempty,max=128"`
	Message string `json:"message" binding:"required,min=10,max=2000"`
	Source  string `json:"source" binding:"omitempty,max=64"`
}


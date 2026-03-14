package utils

import "github.com/gin-gonic/gin"

// Envelope is the common API response format.
type Envelope struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId"`
}

// RespondOK writes a success response.
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(200, Envelope{
		Code:      0,
		Message:   "成功",
		Data:      data,
		RequestID: RequestIDFromContext(c),
	})
}

// RespondCreated writes a created response.
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(201, Envelope{
		Code:      0,
		Message:   "成功",
		Data:      data,
		RequestID: RequestIDFromContext(c),
	})
}

// RespondError writes an error response.
func RespondError(c *gin.Context, err error) {
	appErr := ResolveError(err)
	c.JSON(appErr.StatusCode, Envelope{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: RequestIDFromContext(c),
	})
}

// RequestIDFromContext reads the request ID from context.
func RequestIDFromContext(c *gin.Context) string {
	value, exists := c.Get("requestID")
	if !exists {
		return ""
	}
	requestID, _ := value.(string)
	return requestID
}

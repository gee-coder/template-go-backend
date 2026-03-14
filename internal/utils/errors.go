package utils

import (
	"errors"
	"net/http"
)

var (
	// ErrUnauthorized means the request is unauthenticated.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden means the request has no permission.
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound means the resource was not found.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict means the resource already exists.
	ErrConflict = errors.New("resource already exists")
	// ErrInvalidCredential means the username or password is invalid.
	ErrInvalidCredential = errors.New("invalid username or password")
)

// AppError is a structured business error.
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates an application error.
func NewAppError(code int, statusCode int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// ResolveError converts any error into a business error.
func ResolveError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	switch {
	case errors.Is(err, ErrUnauthorized):
		return NewAppError(http.StatusUnauthorized, http.StatusUnauthorized, "未登录或登录已失效")
	case errors.Is(err, ErrForbidden):
		return NewAppError(http.StatusForbidden, http.StatusForbidden, "没有访问权限")
	case errors.Is(err, ErrNotFound):
		return NewAppError(http.StatusNotFound, http.StatusNotFound, "资源不存在")
	case errors.Is(err, ErrConflict):
		return NewAppError(http.StatusConflict, http.StatusConflict, "资源已存在")
	case errors.Is(err, ErrInvalidCredential):
		return NewAppError(http.StatusUnauthorized, http.StatusUnauthorized, "用户名或密码错误")
	default:
		return NewAppError(http.StatusInternalServerError, http.StatusInternalServerError, "服务器繁忙，请稍后重试")
	}
}


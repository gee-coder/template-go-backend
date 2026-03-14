package middleware

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// PermissionResolver resolves permissions for the current user.
type PermissionResolver func(c *gin.Context, userID uint) ([]string, error)

// PermissionGuard validates the required permission code.
func PermissionGuard(resolver PermissionResolver, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if permission == "" {
			c.Next()
			return
		}

		userID := MustUserID(c)
		permissions, err := resolver(c, userID)
		if err != nil {
			utils.RespondError(c, err)
			c.Abort()
			return
		}

		if !containsPermission(context.Background(), permissions, permission) {
			utils.RespondError(c, utils.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func containsPermission(ctx context.Context, permissions []string, target string) bool {
	_ = ctx
	for _, permission := range permissions {
		if permission == target {
			return true
		}
	}
	return false
}


package middleware

import (
	"net/http"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuth verifies the access token.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.RespondError(c, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.RespondError(c, utils.NewAppError(http.StatusUnauthorized, http.StatusUnauthorized, "Authorization 格式错误"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(secret, parts[1])
		if err != nil {
			utils.RespondError(c, utils.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

// MustUserID returns the current user ID from claims.
func MustUserID(c *gin.Context) uint {
	value, exists := c.Get("claims")
	if !exists {
		return 0
	}

	claims, ok := value.(*utils.TokenClaims)
	if !ok {
		return 0
	}
	return claims.UserID
}


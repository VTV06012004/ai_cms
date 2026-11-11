package middleware

import (
	"strings"

	"github.com/VTV06012004/ai_cms/internal/delivery/http/httperrors"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(credentials *security.Credentials) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httperrors.WriteError(c, httperrors.ErrMissingToken)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			httperrors.WriteError(c, httperrors.ErrMalformedToken)
			c.Abort()
			return
		}

		claims, err := credentials.ValidateToken(tokenString)
		if err != nil {
			httperrors.WriteError(c, security.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func OptionalAuthMiddleware(credentials *security.Credentials) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			httperrors.WriteError(c, httperrors.ErrMalformedToken)
			c.Abort()
			return
		}

		claims, err := credentials.ValidateToken(tokenString)
		if err != nil {
			httperrors.WriteError(c, security.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			httperrors.WriteError(c, httperrors.ErrMissingRoleIdentity)
			c.Abort()
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			httperrors.WriteError(c, httperrors.ErrInvalidRole)
			c.Abort()
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}
		httperrors.WriteError(c, httperrors.ErrForbidden)
		c.Abort()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return RequireRoles("admin")
}

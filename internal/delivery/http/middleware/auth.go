package middleware

import (
	"net/http"
	"strings"

	"api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware kiểm tra tính hợp lệ của JWT token từ Header Authorization
func AuthMiddleware(credentials *security.Credentials) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy token xác thực"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng token không hợp lệ (Bearer <token>)"})
			c.Abort()
			return
		}

		claims, err := credentials.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng token không hợp lệ (******"})
			c.Abort()
			return
		}

		credentials := security.NewCredentials()
		claims, err := credentials.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRoles kiểm tra xem người dùng có thuộc nhóm quyền cho phép hay không
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
			c.Abort()
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Vai trò người dùng không hợp lệ"})
			c.Abort()
			return
		}

		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện hành động này"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnlyMiddleware giới hạn quyền truy cập chỉ dành riêng cho Admin
func AdminOnlyMiddleware() gin.HandlerFunc {
	return RequireRoles("admin")
}

package middleware

import (
	"net/http"
	"strings"

	"api/utils"

	"github.com/gin-gonic/gin"
)

// Middleware kiểm tra JWT Authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu cung cấp Token xác thực"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng Token phải là Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Lưu thông tin User vào context của Gin
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// Middleware kiểm tra Quyền hạn (Role-Based Access Control)
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện thao tác này"})
		c.Abort()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
			c.Abort()
			return
		}

		// Kiểm tra role có phải admin không (chuyển về chữ thường để so sánh an toàn)
		if roleStr, ok := role.(string); !ok || roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Yêu cầu quyền quản trị viên (Admin)!"})
			c.Abort()
			return
		}

		c.Next()
	}
}

package handlers

import (
	"net/http"
	"strconv"

	"api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllUsers - Lấy danh sách toàn bộ người dùng (Dành cho Admin)
func GetAllUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		if err := db.Select("id", "email", "role", "created_at").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách người dùng"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": users})
	}
}

// UpdateUserRole - Cập nhật quyền (Role) cho người dùng
func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		targetUserID, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
			return
		}

		// 1. Lấy ID của chính Admin đang thao tác từ Token (thông qua AuthMiddleware)
		currentUserID, _ := c.Get("user_id")

		// Chặn tuyệt đối nếu tự đổi quyền chính mình
		if currentUserID != nil {
			// Ép kiểu tùy thuộc vào định nghĩa user_id trong claims của bạn (thường là uint hoặc float64)
			var currentID uint
			switch v := currentUserID.(type) {
			case uint:
				currentID = v
			case float64:
				currentID = uint(v)
			case int:
				currentID = uint(v)
			}

			if currentID == uint(targetUserID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không thể tự thay đổi quyền của chính tài khoản mình!"})
				return
			}
		}

		var req struct {
			Role string `json:"role" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
			return
		}

		// Kiểm tra các role hợp lệ
		validRoles := map[string]bool{"admin": true, "editor": true, "author": true}
		if !validRoles[req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Vai trò (Role) không hợp lệ"})
			return
		}

		// Lấy thông tin user mục tiêu từ DB
		var targetUser models.User
		if err := db.First(&targetUser, targetUserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
			return
		}

		// 2. CHẶN TUYỆT ĐỐI: Không ai được phép thay đổi quyền của Admin gốc (ID = 1)
		if targetUser.ID == 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Không được phép thay đổi quyền của Quản trị viên gốc (Root Admin)!"})
			return
		}

		// 3. Thực hiện cập nhật quyền cho các tài khoản thông thường / admin phụ khác
		if err := db.Model(&models.User{}).Where("id = ?", targetUserID).Update("role", req.Role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật vai trò thất bại"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cập nhật vai trò thành công"})
	}
}

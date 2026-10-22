package utils

import (
	"api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RecordAuditLog giúp ghi lại hoạt động người dùng vào cơ sở dữ liệu
func RecordAuditLog(db *gorm.DB, c *gin.Context, action string, target string, details string) {
	var userID uint = 0

	// Hỗ trợ quét nhiều kiểu key khác nhau trong Gin Context để lấy đúng ID
	keys := []string{"user_id", "userID", "id"}
	for _, key := range keys {
		if userIDVal, exists := c.Get(key); exists {
			if id, ok := userIDVal.(uint); ok {
				userID = id
				break
			} else if idInt, ok := userIDVal.(int); ok {
				userID = uint(idInt)
				break
			} else if idFloat, ok := userIDVal.(float64); ok {
				userID = uint(idFloat)
				break
			}
		}
	}

	// Lấy địa chỉ IP người dùng
	ipAddress := c.ClientIP()
	// Tạo đối tượng AuditLog
	logEntry := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Target:    target,
		Details:   details,
		IPAddress: ipAddress,
	}

	// Lưu vào MySQL
	db.Create(&logEntry)
}

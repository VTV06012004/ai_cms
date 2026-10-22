package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
)

type AuditHandler struct {
	DB *gorm.DB
}

func NewAuditHandler(db *gorm.DB) *AuditHandler {
	return &AuditHandler{DB: db}
}

// GetAuditLogs lấy danh sách toàn bộ nhật ký hệ thống (Chỉ dành cho Admin)
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog

	// Query lấy logs kèm thông tin User (Preload), sắp xếp từ mới nhất đến cũ nhất
	if err := h.DB.Preload("User").Order("created_at DESC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách nhật ký hệ thống"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

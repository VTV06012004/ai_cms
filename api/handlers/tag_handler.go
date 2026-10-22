package handlers

import (
	"api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagHandler struct {
	DB *gorm.DB
}

func NewTagHandler(db *gorm.DB) *TagHandler {
	return &TagHandler{DB: db}
}

func (h *TagHandler) GetAll(c *gin.Context) {
	var tags []models.Tag
	if err := h.DB.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách thẻ"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

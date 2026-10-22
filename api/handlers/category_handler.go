package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
)

type CategoryHandler struct {
	DB *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{DB: db}
}

// Lấy danh sách Categories
func (h *CategoryHandler) GetAll(c *gin.Context) {
	var categories []models.Category
	h.DB.Find(&categories)
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// Tạo mới Category
func (h *CategoryHandler) Create(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh mục hoặc Slug đã tồn tại"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

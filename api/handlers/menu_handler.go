package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
	"api/utils"
)

type MenuHandler struct {
	DB *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{DB: db}
}

type MenuItemInput struct {
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type"`
	TargetID  uint   `json:"target_id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type MenuInput struct {
	Name      string          `json:"name" binding:"required"`
	Location  string          `json:"location"`
	MenuItems []MenuItemInput `json:"menu_items"`
}

// Lấy danh sách Menu kèm các items bên trong
func (h *MenuHandler) GetAll(c *gin.Context) {
	var menus []models.Menu
	if err := h.DB.Preload("MenuItems").Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": menus})
}

// Tạo mới Menu
func (h *MenuHandler) Create(c *gin.Context) {
	var input MenuInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	menu := models.Menu{
		Name:     input.Name,
		Location: input.Location,
	}

	if err := h.DB.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo menu"})
		return
	}

	// Thêm các items vào menu
	for _, itemInput := range input.MenuItems {
		url := itemInput.URL
		// Nếu là category, có thể tự động sinh URL hoặc giữ nguyên target_id
		if itemInput.Type == "category" && itemInput.TargetID > 0 {
			var cat models.Category
			if h.DB.First(&cat, itemInput.TargetID).Error == nil {
				if itemInput.Title == "" {
					itemInput.Title = cat.Name
				}
				url = fmt.Sprintf("/category/%d", cat.ID)
			}
		}

		item := models.MenuItem{
			MenuID:    menu.ID,
			Title:     itemInput.Title,
			Type:      itemInput.Type,
			TargetID:  itemInput.TargetID,
			URL:       url,
			SortOrder: itemInput.SortOrder,
		}
		h.DB.Create(&item)
	}

	utils.RecordAuditLog(h.DB, c, "CREATE_MENU", fmt.Sprintf("Menu ID: %d", menu.ID), fmt.Sprintf("Đã tạo menu: '%s'", menu.Name))

	h.DB.Preload("MenuItems").First(&menu, menu.ID)
	c.JSON(http.StatusCreated, gin.H{"data": menu})
}

// Xóa Menu
func (h *MenuHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var menu models.Menu
	if err := h.DB.First(&menu, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy menu"})
		return
	}

	h.DB.Delete(&menu)
	utils.RecordAuditLog(h.DB, c, "DELETE_MENU", fmt.Sprintf("Menu ID: %s", id), fmt.Sprintf("Đã xóa menu: '%s'", menu.Name))

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa menu thành công"})
}

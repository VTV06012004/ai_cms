package models

import (
	"gorm.io/gorm"
)

type Menu struct {
	gorm.Model
	Name      string     `json:"name" binding:"required"` // Tên menu (VD: Menu chính Header)
	Location  string     `json:"location"`                // Vị trí: header, footer
	MenuItems []MenuItem `json:"menu_items" gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;"`
}

type MenuItem struct {
	gorm.Model
	MenuID    uint   `json:"menu_id"`
	Title     string `json:"title" binding:"required"` // Nhãn hiển thị trên menu
	Type      string `json:"type"`                     // Loại: 'category', 'custom'
	TargetID  uint   `json:"target_id"`                // ID của Category (nếu type = category)
	URL       string `json:"url"`                      // Link tùy chỉnh hoặc tự động sinh theo category
	SortOrder int    `json:"sort_order"`               // Thứ tự sắp xếp
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type Revision struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PostID     uint           `gorm:"not null" json:"post_id"`
	Title      string         `gorm:"not null" json:"title"`
	Slug       string         `gorm:"not null" json:"slug"`
	Content    string         `gorm:"type:text" json:"content"`
	Status     string         `json:"status"`
	CategoryID uint           `json:"category_id"`
	UserID     uint           `gorm:"not null" json:"user_id"` // Người thực hiện chỉnh sửa
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

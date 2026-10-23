package models

import (
	"time"

	"gorm.io/gorm"
)

// User - Bảng quản lý người dùng & phân quyền (RBAC)
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"unique;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Name         string         `json:"name"`
	Role         string         `gorm:"type:enum('admin','editor','author');default:'author'" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category - Bảng danh mục bài viết (2 cấp)
type Category struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"not null" json:"name"`
	Slug     string `gorm:"size:255;unique;not null" json:"slug"`
	ParentID *uint  `json:"parent_id"`
}

// Post - Bảng quản lý nội dung bài viết CMS
type Post struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Title           string         `gorm:"not null" json:"title"`
	Slug            string         `json:"slug" gorm:"size:255;uniqueIndex:idx_posts_slug_deleted;not null"`
	Content         string         `gorm:"type:longtext" json:"content"`
	Excerpt         string         `gorm:"type:text" json:"excerpt"`
	Status          string         `gorm:"type:enum('draft','review','approved','rejected','published');default:'draft'" json:"status"`
	RejectionReason string         `gorm:"type:text" json:"rejection_reason"`
	AuthorID        uint           `json:"author_id"`
	Author          User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	ModeratorID     *uint          `json:"moderator_id"`
	Moderator       *User          `gorm:"foreignKey:ModeratorID" json:"moderator,omitempty"`
	CategoryID      uint           `json:"category_id"`
	Category        Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	SEOTitle        string         `json:"seo_title"`
	SEODescription  string         `json:"seo_description"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index;uniqueIndex:idx_posts_slug_deleted"`
	Tags            []Tag          `json:"tags" gorm:"many2many:post_tags;"`
}

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"` // Dùng ID chuẩn thay vì gorm.Model để đồng bộ với Category/User
	Name      string         `json:"name" gorm:"unique;not null;size:255"`
	Slug      string         `json:"slug" gorm:"unique;not null;size:255"`
	Posts     []Post         `json:"posts,omitempty" gorm:"many2many:post_tags;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

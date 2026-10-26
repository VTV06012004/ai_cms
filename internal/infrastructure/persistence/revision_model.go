package persistence

import (
	"api/internal/domain"
	"time"

	"gorm.io/gorm"
)

// RevisionModel đại diện cho bảng revisions trong cơ sở dữ liệu
type RevisionModel struct {
	ID         uint      `gorm:"primaryKey"`
	PostID     uint      `gorm:"index"`
	Title      string    `gorm:"not null"`
	Slug       string    `gorm:"not null"`
	Content    string    `gorm:"type:text"`
	Status     string    `gorm:"default:'draft'"`
	CategoryID uint      `gorm:"index"`
	UserID     uint      `gorm:"index"`
	User       UserModel `gorm:"foreignKey:UserID"` // Phục vụ Preload thông tin người chỉnh sửa phiên bản
	CreatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (RevisionModel) TableName() string {
	return "revisions"
}

func (m RevisionModel) ToDomain() domain.Revision {
	data := domain.RevisionData{
		ID:         m.ID,
		PostID:     m.PostID,
		Title:      m.Title,
		Slug:       m.Slug,
		Content:    m.Content,
		Status:     domain.PostStatus(m.Status),
		CategoryID: m.CategoryID,
		UserID:     m.UserID,
		CreatedAt:  m.CreatedAt,
	}
	if m.User.ID != 0 {
		data.User = userModelToDomain(m.User)
	}
	return *domain.ReconstituteRevision(data)
}

func revisionModelFromDomain(revision *domain.Revision) RevisionModel {
	data := revision.Data()
	return RevisionModel{
		ID:         data.ID,
		PostID:     data.PostID,
		Title:      data.Title,
		Slug:       data.Slug,
		Content:    data.Content,
		Status:     string(data.Status),
		CategoryID: data.CategoryID,
		UserID:     data.UserID,
		CreatedAt:  data.CreatedAt,
	}
}

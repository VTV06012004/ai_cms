package persistence

import (
	"time"

	"github.com/VTV06012004/ai_cms/internal/domain"

	"gorm.io/gorm"
)

type PostModel struct {
	ID              uint      `gorm:"primaryKey"`
	Title           string    `gorm:"not null"`
	Slug            string    `gorm:"size:255;not null"`
	Content         string    `gorm:"type:longtext"`
	Status          string    `gorm:"type:enum('draft','review','rejected','published','scheduled');default:'draft'"`
	RejectionReason string    `gorm:"type:text"`
	AuthorID        uint      `gorm:"index"`
	Author          UserModel `gorm:"foreignKey:AuthorID"`
	ModeratorID     *uint
	Moderator       *UserModel    `gorm:"foreignKey:ModeratorID"`
	CategoryID      uint          `gorm:"index"`
	Category        CategoryModel `gorm:"foreignKey:CategoryID"`
	ScheduledAt     *time.Time    `gorm:"index"`
	PublishedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	Tags            []TagModel     `gorm:"many2many:post_tags;"`
}

func (PostModel) TableName() string {
	return "posts"
}

func (m PostModel) ToDomain() domain.Post {
	data := domain.PostData{
		ID:              m.ID,
		Title:           m.Title,
		Slug:            m.Slug,
		Content:         m.Content,
		Status:          domain.PostStatus(m.Status),
		RejectionReason: m.RejectionReason,
		AuthorID:        m.AuthorID,
		ModeratorID:     m.ModeratorID,
		CategoryID:      m.CategoryID,
		PublishedAt:     m.PublishedAt,
		ScheduledAt:     m.ScheduledAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
	if m.Author.ID != 0 {
		data.Author = userModelToDomain(m.Author)
	}
	if m.Moderator != nil && m.Moderator.ID != 0 {
		data.Moderator = userModelToDomain(*m.Moderator)
	}
	if m.Category.ID != 0 {
		category := m.Category.ToDomain()
		data.Category = &category
	}
	data.Tags = make([]domain.Tag, 0, len(m.Tags))
	for _, tag := range m.Tags {
		data.Tags = append(data.Tags, tag.ToDomain())
	}
	return *domain.ReconstitutePost(data)
}

func postModelFromDomain(post *domain.Post) PostModel {
	data := post.Data()
	return PostModel{
		ID:              data.ID,
		Title:           data.Title,
		Slug:            data.Slug,
		Content:         data.Content,
		Status:          string(data.Status),
		RejectionReason: data.RejectionReason,
		AuthorID:        data.AuthorID,
		ModeratorID:     data.ModeratorID,
		CategoryID:      data.CategoryID,
		PublishedAt:     data.PublishedAt,
		ScheduledAt:     data.ScheduledAt,
		CreatedAt:       data.CreatedAt,
		UpdatedAt:       data.UpdatedAt,
	}
}

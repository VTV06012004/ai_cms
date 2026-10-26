package persistence

import (
	"time"

	"api/internal/domain"

	"gorm.io/gorm"
)

type TagModel struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null;size:255"`
	Slug      string `gorm:"unique;not null;size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (TagModel) TableName() string { return "tags" }

func (m TagModel) ToDomain() domain.Tag {
	return domain.ReconstituteTag(domain.TagData{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	})
}

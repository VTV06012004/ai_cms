package persistence

import "api/internal/domain"

type CategoryModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Slug        string `gorm:"size:255;unique;not null"`
	Description string
	ParentID    *uint
}

func (CategoryModel) TableName() string { return "categories" }

func (m CategoryModel) ToDomain() domain.Category {
	return *domain.ReconstituteCategory(domain.CategoryData{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
		ParentID:    m.ParentID,
	})
}

func FromDomainCategory(d domain.Category) CategoryModel {
	data := d.Data()
	return CategoryModel{
		ID:          data.ID,
		Name:        data.Name,
		Slug:        data.Slug,
		Description: data.Description,
		ParentID:    data.ParentID,
	}
}

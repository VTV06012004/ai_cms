package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) domain.TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) GetAll() ([]domain.Tag, error) {
	var models []TagModel
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Tag
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

var _ domain.TagRepository = (*TagRepository)(nil)

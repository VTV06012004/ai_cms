package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll() ([]domain.Category, error) {
	var models []CategoryModel
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Category
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

func (r *CategoryRepository) GetByID(id uint) (*domain.Category, error) {
	var model CategoryModel
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}
	category := model.ToDomain()
	return &category, nil
}

func (r *CategoryRepository) Create(category *domain.Category) (*domain.Category, error) {
	model := FromDomainCategory(*category)
	if err := r.db.Create(&model).Error; err != nil {
		return nil, err
	}
	created := model.ToDomain()
	return &created, nil
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type RevisionRepository struct {
	db *gorm.DB
}

func NewRevisionRepository(db *gorm.DB) domain.RevisionRepository {
	return &RevisionRepository{db: db}
}

func (r *RevisionRepository) Create(revision *domain.Revision) error {
	model := revisionModelFromDomain(revision)
	return r.db.Create(&model).Error
}

func (r *RevisionRepository) GetByPostID(postID uint) ([]domain.Revision, error) {
	var models []RevisionModel
	if err := r.db.Preload("User").Where("post_id = ?", postID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Revision, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDomain())
	}
	return result, nil
}

func (r *RevisionRepository) GetByID(id uint) (*domain.Revision, error) {
	var model RevisionModel
	if err := r.db.Preload("User").First(&model, id).Error; err != nil {
		return nil, err
	}

	var _ domain.RevisionRepository = (*RevisionRepository)(nil)
	revision := model.ToDomain()
	return &revision, nil
}

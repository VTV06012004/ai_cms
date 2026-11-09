package persistence

import (
	"api/internal/domain"
	"context"

	"gorm.io/gorm"
)

type RevisionRepository struct {
	db *gorm.DB
}

func NewRevisionRepository(db *gorm.DB) domain.RevisionRepository {
	return &RevisionRepository{db: db}
}

func (r *RevisionRepository) Create(ctx context.Context, revision *domain.Revision) error {
	model := revisionModelFromDomain(revision)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *RevisionRepository) GetByPostID(ctx context.Context, postID uint) ([]domain.Revision, error) {
	var models []RevisionModel
	if err := r.db.WithContext(ctx).Preload("User").Where("post_id = ?", postID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Revision, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDomain())
	}
	return result, nil
}

func (r *RevisionRepository) GetByID(ctx context.Context, id uint) (*domain.Revision, error) {
	var model RevisionModel
	if err := r.db.WithContext(ctx).Preload("User").First(&model, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrRevisionNotFound)
	}

	var _ domain.RevisionRepository = (*RevisionRepository)(nil)
	revision := model.ToDomain()
	return &revision, nil
}

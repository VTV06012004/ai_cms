package persistence

import (
	"api/internal/domain"
	"context"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) domain.TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) GetAll(ctx context.Context) ([]domain.Tag, error) {
	var models []TagModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Tag
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

func (r *TagRepository) GetByID(ctx context.Context, id uint) (*domain.Tag, error) {
	var model TagModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrTagNotFound)
	}
	tag := model.ToDomain()
	return &tag, nil
}

func (r *TagRepository) Create(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	data := tag.Data()
	model := TagModel{Name: data.Name, Slug: data.Slug}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateKey(err, "tags") {
			return domain.Tag{}, domain.ErrTagAlreadyExists
		}
		return domain.Tag{}, err
	}
	return model.ToDomain(), nil
}

func (r *TagRepository) Update(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	data := tag.Data()
	err := r.db.WithContext(ctx).Model(&TagModel{}).Where("id = ?", data.ID).
		Updates(map[string]any{"name": data.Name, "slug": data.Slug}).Error
	if err != nil {
		if isDuplicateKey(err, "tags") {
			return domain.Tag{}, domain.ErrTagAlreadyExists
		}
		return domain.Tag{}, err
	}
	updated, err := r.GetByID(ctx, data.ID)
	if err != nil {
		return domain.Tag{}, err
	}
	return *updated, nil
}

// Delete xóa hẳn thẻ (kể cả khóa slug để tạo lại được) và gỡ thẻ khỏi các bài viết.
func (r *TagRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM post_tags WHERE tag_model_id = ?", id).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Delete(&TagModel{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrTagNotFound
		}
		return nil
	})
}

var _ domain.TagRepository = (*TagRepository)(nil)

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
	if err := r.db.Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Tag
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

func (r *TagRepository) GetByID(id uint) (*domain.Tag, error) {
	var model TagModel
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrTagNotFound)
	}
	tag := model.ToDomain()
	return &tag, nil
}

func (r *TagRepository) Create(tag domain.Tag) (domain.Tag, error) {
	data := tag.Data()
	model := TagModel{Name: data.Name, Slug: data.Slug}
	if err := r.db.Create(&model).Error; err != nil {
		if isDuplicateKey(err, "tags") {
			return domain.Tag{}, domain.ErrTagAlreadyExists
		}
		return domain.Tag{}, err
	}
	return model.ToDomain(), nil
}

func (r *TagRepository) Update(tag domain.Tag) (domain.Tag, error) {
	data := tag.Data()
	err := r.db.Model(&TagModel{}).Where("id = ?", data.ID).
		Updates(map[string]any{"name": data.Name, "slug": data.Slug}).Error
	if err != nil {
		if isDuplicateKey(err, "tags") {
			return domain.Tag{}, domain.ErrTagAlreadyExists
		}
		return domain.Tag{}, err
	}
	updated, err := r.GetByID(data.ID)
	if err != nil {
		return domain.Tag{}, err
	}
	return *updated, nil
}

// Delete xóa hẳn thẻ (kể cả khóa slug để tạo lại được) và gỡ thẻ khỏi các bài viết.
func (r *TagRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

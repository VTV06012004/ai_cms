package persistence

import (
	"api/internal/domain"
	"context"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	var models []CategoryModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Category
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	var model CategoryModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrCategoryNotFound)
	}
	category := model.ToDomain()
	return &category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	model := FromDomainCategory(*category)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateKey(err, "categories_slug") {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	created := model.ToDomain()
	return &created, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	model := FromDomainCategory(*category)
	err := r.db.WithContext(ctx).Model(&CategoryModel{}).Where("id = ?", model.ID).
		Select("name", "slug", "description", "parent_id").
		Updates(&model).Error
	if err != nil {
		if isDuplicateKey(err, "categories_slug") {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return r.GetByID(ctx, model.ID)
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&CategoryModel{}, id).Error
	if isForeignKeyViolation(err) {
		return domain.ErrCategoryInUse
	}
	return err
}

// CountPosts đếm cả bài đã xóa mềm, vì chúng vẫn tham chiếu danh mục trong DB.
func (r *CategoryRepository) CountPosts(ctx context.Context, id uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&PostModel{}).Where("category_id = ?", id).Count(&count).Error
	return count, err
}

func (r *CategoryRepository) CountChildren(ctx context.Context, id uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CategoryModel{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

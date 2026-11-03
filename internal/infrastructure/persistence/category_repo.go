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
		return nil, notFoundOr(err, domain.ErrCategoryNotFound)
	}
	category := model.ToDomain()
	return &category, nil
}

func (r *CategoryRepository) Create(category *domain.Category) (*domain.Category, error) {
	model := FromDomainCategory(*category)
	if err := r.db.Create(&model).Error; err != nil {
		if isDuplicateKey(err, "categories_slug") {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	created := model.ToDomain()
	return &created, nil
}

func (r *CategoryRepository) Update(category *domain.Category) (*domain.Category, error) {
	model := FromDomainCategory(*category)
	err := r.db.Model(&CategoryModel{}).Where("id = ?", model.ID).
		Select("name", "slug", "description", "parent_id").
		Updates(&model).Error
	if err != nil {
		if isDuplicateKey(err, "categories_slug") {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, err
	}
	return r.GetByID(model.ID)
}

func (r *CategoryRepository) Delete(id uint) error {
	err := r.db.Delete(&CategoryModel{}, id).Error
	if isForeignKeyViolation(err) {
		return domain.ErrCategoryInUse
	}
	return err
}

// CountPosts đếm cả bài đã xóa mềm, vì chúng vẫn tham chiếu danh mục trong DB.
func (r *CategoryRepository) CountPosts(id uint) (int64, error) {
	var count int64
	err := r.db.Unscoped().Model(&PostModel{}).Where("category_id = ?", id).Count(&count).Error
	return count, err
}

func (r *CategoryRepository) CountChildren(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&CategoryModel{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

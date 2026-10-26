package usecase

import (
	"errors"

	"api/internal/domain"
)

type CategoryUseCase struct {
	categoryRepo domain.CategoryRepository
}

func NewCategoryUseCase(repo domain.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{categoryRepo: repo}
}

func (uc *CategoryUseCase) GetAllCategories() ([]domain.Category, error) {
	return uc.categoryRepo.GetAll()
}

func (uc *CategoryUseCase) CreateCategory(name, slug string, parentID *uint) (*domain.Category, error) {
	category, err := domain.NewCategory(name, slug, parentID)
	if err != nil {
		return nil, err
	}
	created, err := uc.categoryRepo.Create(category)
	if err != nil {
		return nil, errors.New("danh mục hoặc slug đã tồn tại")
	}
	return created, nil
}

package usecase

import (
	"errors"
	"fmt"

	"api/internal/domain"
)

const maxCategoryDepth = 50

type CategoryUseCase struct {
	categoryRepo domain.CategoryRepository
	unitOfWork   UnitOfWork
}

func NewCategoryUseCase(repo domain.CategoryRepository, unitOfWork UnitOfWork) *CategoryUseCase {
	return &CategoryUseCase{categoryRepo: repo, unitOfWork: unitOfWork}
}

type CategoryInputDTO struct {
	Name        string
	Slug        string
	Description string
	ParentID    *uint
}

func (uc *CategoryUseCase) GetAllCategories() ([]domain.Category, error) {
	return uc.categoryRepo.GetAll()
}

func (uc *CategoryUseCase) CreateCategory(userID uint, input CategoryInputDTO, ipAddress string) (*domain.Category, error) {
	category, err := domain.NewCategory(input.Name, input.Slug, input.ParentID)
	if err != nil {
		return nil, err
	}
	category.Describe(input.Description)

	var created *domain.Category
	err = uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		if err := ensureValidCategoryParent(repositories.Categories, 0, input.ParentID); err != nil {
			return err
		}
		saved, err := repositories.Categories.Create(category)
		if err != nil {
			return err
		}
		created = saved
		data := saved.Data()
		return recordCatalogAudit(repositories, userID, "CREATE_CATEGORY",
			fmt.Sprintf("Danh mục ID: %d", data.ID), fmt.Sprintf("Đã tạo danh mục: '%s'", data.Name), ipAddress)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (uc *CategoryUseCase) UpdateCategory(userID, id uint, input CategoryInputDTO, ipAddress string) (*domain.Category, error) {
	var updated *domain.Category
	err := uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		category, err := repositories.Categories.GetByID(id)
		if err != nil {
			return err
		}
		if err := ensureValidCategoryParent(repositories.Categories, id, input.ParentID); err != nil {
			return err
		}
		if err := category.Update(input.Name, input.Slug, input.Description, input.ParentID); err != nil {
			return err
		}
		saved, err := repositories.Categories.Update(category)
		if err != nil {
			return err
		}
		updated = saved
		return recordCatalogAudit(repositories, userID, "UPDATE_CATEGORY",
			fmt.Sprintf("Danh mục ID: %d", id), fmt.Sprintf("Đã cập nhật danh mục: '%s'", saved.Data().Name), ipAddress)
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (uc *CategoryUseCase) DeleteCategory(userID, id uint, ipAddress string) error {
	return uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		category, err := repositories.Categories.GetByID(id)
		if err != nil {
			return err
		}
		posts, err := repositories.Categories.CountPosts(id)
		if err != nil {
			return err
		}
		if posts > 0 {
			return domain.ErrCategoryInUse
		}
		children, err := repositories.Categories.CountChildren(id)
		if err != nil {
			return err
		}
		if children > 0 {
			return domain.ErrCategoryHasChildren
		}
		if err := repositories.Categories.Delete(id); err != nil {
			return err
		}
		return recordCatalogAudit(repositories, userID, "DELETE_CATEGORY",
			fmt.Sprintf("Danh mục ID: %d", id), fmt.Sprintf("Đã xóa danh mục: '%s'", category.Data().Name), ipAddress)
	})
}

// ensureValidCategoryParent là nơi duy nhất kiểm tra danh mục cha: phải tồn tại,
// không được là chính nó hay con cháu của nó. id = 0 khi đang tạo mới.
func ensureValidCategoryParent(categories domain.CategoryRepository, id uint, parentID *uint) error {
	if parentID == nil {
		return nil
	}
	current := *parentID
	for depth := 0; depth < maxCategoryDepth; depth++ {
		if current == id {
			return domain.ErrInvalidCategoryParent
		}
		parent, err := categories.GetByID(current)
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return domain.ErrInvalidCategoryParent
		}
		if err != nil {
			return err
		}
		next := parent.Data().ParentID
		if next == nil {
			return nil
		}
		current = *next
	}
	return domain.ErrInvalidCategoryParent
}

func recordCatalogAudit(repositories TransactionRepositories, userID uint, action, target, details, ipAddress string) error {
	if err := repositories.Audits.RecordLog(userID, action, target, details, ipAddress); err != nil {
		return fmt.Errorf("không thể ghi audit log: %w", err)
	}
	return nil
}

package usecase

import (
	"fmt"

	"api/internal/domain"
)

type MenuUseCase struct {
	menuRepo     domain.MenuRepository
	categoryRepo domain.CategoryRepository
	unitOfWork   UnitOfWork
}

func NewMenuUseCase(menus domain.MenuRepository, categories domain.CategoryRepository, unitOfWork UnitOfWork) *MenuUseCase {
	return &MenuUseCase{menuRepo: menus, categoryRepo: categories, unitOfWork: unitOfWork}
}

type MenuItemInputDTO struct {
	Title     string
	Type      string
	TargetID  uint
	URL       string
	SortOrder int
}

func (uc *MenuUseCase) GetAllMenus() ([]domain.Menu, error) {
	return uc.menuRepo.GetAll()
}

type MenuInputDTO struct {
	Name      string
	Location  string
	MenuItems []MenuItemInputDTO
}

func (uc *MenuUseCase) CreateMenu(userID uint, input MenuInputDTO, ipAddress string) (*domain.Menu, error) {
	menu, err := domain.NewMenu(input.Name, input.Location)
	if err != nil {
		return nil, err
	}
	for _, itemInput := range input.MenuItems {
		url := itemInput.URL
		title := itemInput.Title
		if itemInput.Type == "category" && itemInput.TargetID > 0 {
			if category, err := uc.categoryRepo.GetByID(itemInput.TargetID); err == nil {
				categoryData := category.Data()
				if title == "" {
					title = categoryData.Name
				}
				url = fmt.Sprintf("/category/%d", categoryData.ID)
			}
		}
		item, err := domain.NewMenuItem(title, itemInput.Type, itemInput.TargetID, url, itemInput.SortOrder)
		if err != nil {
			return nil, err
		}
		menu.AddItem(item)
	}
	var created *domain.Menu
	err = uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		saved, err := repositories.Menus.Create(menu)
		if err != nil {
			return fmt.Errorf("không thể tạo menu: %w", err)
		}
		data := saved.Data()
		if err := repositories.Audits.RecordLog(userID, "CREATE_MENU", fmt.Sprintf("Menu ID: %d", data.ID), fmt.Sprintf("Đã tạo menu: '%s'", data.Name), ipAddress); err != nil {
			return fmt.Errorf("không thể ghi audit log: %w", err)
		}
		created, err = repositories.Menus.GetByID(data.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("không thể hoàn tất tạo menu: %w", err)
	}
	return created, nil
}

func (uc *MenuUseCase) DeleteMenu(userID, menuID uint, ipAddress string) error {
	return uc.unitOfWork.WithinTransaction(func(repositories TransactionRepositories) error {
		menu, err := repositories.Menus.GetByID(menuID)
		if err != nil {
			return fmt.Errorf("không tìm thấy menu: %w", err)
		}
		menuData := menu.Data()
		if err := repositories.Menus.Delete(menuID); err != nil {
			return fmt.Errorf("không thể xóa menu: %w", err)
		}
		return repositories.Audits.RecordLog(userID, "DELETE_MENU", fmt.Sprintf("Menu ID: %d", menuID), fmt.Sprintf("Đã xóa menu: '%s'", menuData.Name), ipAddress)
	})
}

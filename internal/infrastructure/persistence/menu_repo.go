package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) domain.MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetAll() ([]domain.Menu, error) {
	var models []MenuModel
	if err := r.db.Preload("MenuItems").Find(&models).Error; err != nil {
		return nil, err
	}
	var domains []domain.Menu
	for _, m := range models {
		domains = append(domains, m.ToDomain())
	}
	return domains, nil
}

func (r *MenuRepository) GetByID(id uint) (*domain.Menu, error) {
	var model MenuModel
	if err := r.db.Preload("MenuItems").First(&model, id).Error; err != nil {
		return nil, err
	}
	domainMenu := model.ToDomain()
	return &domainMenu, nil
}

func (r *MenuRepository) Create(menu *domain.Menu) (*domain.Menu, error) {
	var created *domain.Menu
	err := r.db.Transaction(func(tx *gorm.DB) error {
		data := menu.Data()
		model := MenuModel{
			Name:     data.Name,
			Location: data.Location,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		model.MenuItems = make([]MenuItemModel, 0, len(data.MenuItems))
		for _, item := range data.MenuItems {
			itemData := item.Data()
			model.MenuItems = append(model.MenuItems, MenuItemModel{
				MenuID:    model.ID,
				Title:     itemData.Title,
				Type:      itemData.Type,
				TargetID:  itemData.TargetID,
				URL:       itemData.URL,
				SortOrder: itemData.SortOrder,
			})
		}
		if len(model.MenuItems) > 0 {
			if err := tx.Create(&model.MenuItems).Error; err != nil {
				return err
			}
		}
		loaded := model.ToDomain()
		created = &loaded
		return nil
	})
	return created, err
}

func (r *MenuRepository) Delete(id uint) error {
	return r.db.Delete(&MenuModel{}, id).Error
}

var _ domain.MenuRepository = (*MenuRepository)(nil)

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
	if err := r.db.Preload("MenuItems", orderMenuItems).Find(&models).Error; err != nil {
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
	if err := r.db.Preload("MenuItems", orderMenuItems).First(&model, id).Error; err != nil {
		return nil, notFoundOr(err, domain.ErrMenuNotFound)
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

func orderMenuItems(db *gorm.DB) *gorm.DB {
	return db.Order("sort_order ASC, id ASC")
}

// Update đổi tên, vị trí và thay toàn bộ các mục của menu trong một transaction.
func (r *MenuRepository) Update(menu *domain.Menu) (*domain.Menu, error) {
	data := menu.Data()
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&MenuModel{}).Where("id = ?", data.ID).
			Updates(map[string]any{"name": data.Name, "location": data.Location}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("menu_id = ?", data.ID).Delete(&MenuItemModel{}).Error; err != nil {
			return err
		}
		items := make([]MenuItemModel, 0, len(data.MenuItems))
		for _, item := range data.MenuItems {
			itemData := item.Data()
			items = append(items, MenuItemModel{
				MenuID: data.ID, Title: itemData.Title, Type: itemData.Type,
				TargetID: itemData.TargetID, URL: itemData.URL, SortOrder: itemData.SortOrder,
			})
		}
		if len(items) > 0 {
			return tx.Create(&items).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(data.ID)
}

var _ domain.MenuRepository = (*MenuRepository)(nil)

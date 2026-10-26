package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type MenuItemModel struct {
	gorm.Model
	MenuID    uint
	Title     string
	Type      string
	TargetID  uint
	URL       string
	SortOrder int
}

func (MenuItemModel) TableName() string { return "menu_items" }

func (m MenuItemModel) ToDomain() domain.MenuItem {
	return domain.ReconstituteMenuItem(domain.MenuItemData{
		ID:        m.ID,
		MenuID:    m.MenuID,
		Title:     m.Title,
		Type:      m.Type,
		TargetID:  m.TargetID,
		URL:       m.URL,
		SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	})
}

type MenuModel struct {
	gorm.Model
	Name      string
	Location  string
	MenuItems []MenuItemModel `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;"`
}

func (MenuModel) TableName() string { return "menus" }

func (m MenuModel) ToDomain() domain.Menu {
	var items []domain.MenuItem
	for _, item := range m.MenuItems {
		items = append(items, item.ToDomain())
	}
	return *domain.ReconstituteMenu(domain.MenuData{
		ID:        m.ID,
		Name:      m.Name,
		Location:  m.Location,
		MenuItems: items,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	})
}

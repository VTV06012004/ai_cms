package domain

import (
	"strings"
	"time"
)

type MenuItem struct {
	id        uint
	menuID    uint
	title     string
	itemType  string
	targetID  uint
	url       string
	sortOrder int
	createdAt time.Time
	updatedAt time.Time
}

type MenuItemData struct {
	ID        uint
	MenuID    uint
	Title     string
	Type      string
	TargetID  uint
	URL       string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ReconstituteMenuItem(data MenuItemData) MenuItem {
	return MenuItem{
		id: data.ID, menuID: data.MenuID, title: data.Title, itemType: data.Type,
		targetID: data.TargetID, url: data.URL, sortOrder: data.SortOrder,
		createdAt: data.CreatedAt, updatedAt: data.UpdatedAt,
	}
}

func NewMenuItem(title, itemType string, targetID uint, url string, sortOrder int) (MenuItem, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return MenuItem{}, ErrMenuItemTitleRequired
	}
	return ReconstituteMenuItem(MenuItemData{
		Title: title, Type: itemType, TargetID: targetID, URL: url, SortOrder: sortOrder,
	}), nil
}

func (i MenuItem) Data() MenuItemData {
	return MenuItemData{
		ID: i.id, MenuID: i.menuID, Title: i.title, Type: i.itemType,
		TargetID: i.targetID, URL: i.url, SortOrder: i.sortOrder,
		CreatedAt: i.createdAt, UpdatedAt: i.updatedAt,
	}
}

type Menu struct {
	id        uint
	name      string
	location  string
	menuItems []MenuItem
	createdAt time.Time
	updatedAt time.Time
}

type MenuData struct {
	ID        uint
	Name      string
	Location  string
	MenuItems []MenuItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMenu(name, location string) (*Menu, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrMenuNameRequired
	}
	return &Menu{name: name, location: location}, nil
}

func ReconstituteMenu(data MenuData) *Menu {
	return &Menu{
		id: data.ID, name: data.Name, location: data.Location,
		menuItems: append([]MenuItem(nil), data.MenuItems...),
		createdAt: data.CreatedAt, updatedAt: data.UpdatedAt,
	}
}

func (m *Menu) AddItem(item MenuItem) {
	m.menuItems = append(m.menuItems, item)
}

func (m *Menu) Data() MenuData {
	return MenuData{
		ID: m.id, Name: m.name, Location: m.location,
		MenuItems: append([]MenuItem(nil), m.menuItems...),
		CreatedAt: m.createdAt, UpdatedAt: m.updatedAt,
	}
}

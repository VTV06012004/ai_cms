package domain

import (
	"errors"
	"testing"
)

func TestNewCategoryRequiresNameAndSlug(t *testing.T) {
	if _, err := NewCategory(" ", "slug", nil); !errors.Is(err, ErrCategoryNameRequired) {
		t.Fatalf("expected ErrCategoryNameRequired, got %v", err)
	}
	if _, err := NewCategory("News", " ", nil); !errors.Is(err, ErrCategorySlugRequired) {
		t.Fatalf("expected ErrCategorySlugRequired, got %v", err)
	}
}

func TestMenuCopiesItemsAndRequiresItemTitle(t *testing.T) {
	menu, err := NewMenu(" Primary ", "header")
	if err != nil {
		t.Fatalf("NewMenu returned error: %v", err)
	}
	item, err := NewMenuItem("Home", "custom", 0, "/", 1)
	if err != nil {
		t.Fatalf("NewMenuItem returned error: %v", err)
	}
	menu.AddItem(item)
	items := menu.Data().MenuItems
	items[0] = MenuItem{}
	if menu.Data().MenuItems[0].Data().Title != "Home" {
		t.Fatal("mutating menu data changed the menu aggregate")
	}
	if _, err := NewMenuItem(" ", "custom", 0, "/", 1); !errors.Is(err, ErrMenuItemTitleRequired) {
		t.Fatalf("expected ErrMenuItemTitleRequired, got %v", err)
	}
}

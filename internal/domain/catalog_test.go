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

func TestCategoryUpdateValidatesAndReplacesFields(t *testing.T) {
	category, err := NewCategory("News", "news", nil)
	if err != nil {
		t.Fatal(err)
	}
	parent := uint(3)
	if err := category.Update("  Tech  ", " tech ", " About tech ", &parent); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	data := category.Data()
	if data.Name != "Tech" || data.Slug != "tech" || data.Description != "About tech" ||
		data.ParentID == nil || *data.ParentID != 3 {
		t.Fatalf("unexpected data after update: %+v", data)
	}

	if err := category.Update(" ", "x", "", nil); !errors.Is(err, ErrCategoryNameRequired) {
		t.Fatalf("expected ErrCategoryNameRequired, got %v", err)
	}
	if category.Data().Name != "Tech" {
		t.Fatal("a failed update changed the category")
	}
}

func TestNewTagAndRename(t *testing.T) {
	if _, err := NewTag(" ", "go"); !errors.Is(err, ErrTagNameRequired) {
		t.Fatalf("expected ErrTagNameRequired, got %v", err)
	}
	if _, err := NewTag("Go", " "); !errors.Is(err, ErrTagSlugRequired) {
		t.Fatalf("expected ErrTagSlugRequired, got %v", err)
	}

	tag := ReconstituteTag(TagData{ID: 4, Name: "Go", Slug: "go"})
	renamed, err := tag.Rename(" Golang ", " golang ")
	if err != nil {
		t.Fatal(err)
	}
	if data := renamed.Data(); data.ID != 4 || data.Name != "Golang" || data.Slug != "golang" {
		t.Fatalf("unexpected renamed tag: %+v", data)
	}
	if tag.Data().Name != "Go" {
		t.Fatal("Rename changed the original tag")
	}
}

func TestMenuRenameAndReplaceItems(t *testing.T) {
	menu, err := NewMenu("Primary", "header")
	if err != nil {
		t.Fatal(err)
	}
	old, _ := NewMenuItem("Old", "custom", 0, "/old", 1)
	menu.AddItem(old)
	fresh, _ := NewMenuItem("New", "custom", 0, "/new", 1)

	menu.ReplaceItems([]MenuItem{fresh})
	if err := menu.Rename("  Footer ", "footer"); err != nil {
		t.Fatal(err)
	}
	data := menu.Data()
	if data.Name != "Footer" || data.Location != "footer" ||
		len(data.MenuItems) != 1 || data.MenuItems[0].Data().Title != "New" {
		t.Fatalf("unexpected menu: %+v", data)
	}
	if err := menu.Rename(" ", "x"); !errors.Is(err, ErrMenuNameRequired) {
		t.Fatalf("expected ErrMenuNameRequired, got %v", err)
	}
}

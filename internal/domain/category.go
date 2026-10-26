package domain

import "strings"

type Category struct {
	id          uint
	name        string
	slug        string
	description string
	parentID    *uint
}

type CategoryData struct {
	ID          uint
	Name        string
	Slug        string
	Description string
	ParentID    *uint
}

func NewCategory(name, slug string, parentID *uint) (*Category, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, ErrCategoryNameRequired
	}
	if slug == "" {
		return nil, ErrCategorySlugRequired
	}
	return ReconstituteCategory(CategoryData{Name: name, Slug: slug, ParentID: parentID}), nil
}

func ReconstituteCategory(data CategoryData) *Category {
	var parentID *uint
	if data.ParentID != nil {
		id := *data.ParentID
		parentID = &id
	}
	return &Category{
		id: data.ID, name: data.Name, slug: data.Slug,
		description: data.Description, parentID: parentID,
	}
}

func (c *Category) Data() CategoryData {
	var parentID *uint
	if c.parentID != nil {
		id := *c.parentID
		parentID = &id
	}
	return CategoryData{
		ID: c.id, Name: c.name, Slug: c.slug,
		Description: c.description, ParentID: parentID,
	}
}

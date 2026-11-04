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

func (c *Category) Describe(description string) {
	c.description = strings.TrimSpace(description)
}

// Update thay toàn bộ thông tin; dữ liệu không hợp lệ thì không đổi gì.
// Quy tắc về danh mục cha (không tự làm cha mình, không vòng lặp) nằm ở usecase vì cần tra cứu DB.
func (c *Category) Update(name, slug, description string, parentID *uint) error {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return ErrCategoryNameRequired
	}
	if slug == "" {
		return ErrCategorySlugRequired
	}
	c.name = name
	c.slug = slug
	c.description = strings.TrimSpace(description)
	c.parentID = nil
	if parentID != nil {
		id := *parentID
		c.parentID = &id
	}
	return nil
}

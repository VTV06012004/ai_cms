package domain

import (
	"strings"
	"time"
)

type Tag struct {
	id        uint
	name      string
	slug      string
	createdAt time.Time
	updatedAt time.Time
}

type TagData struct {
	ID        uint
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ReconstituteTag(data TagData) Tag {
	return Tag{
		id: data.ID, name: data.Name, slug: data.Slug,
		createdAt: data.CreatedAt, updatedAt: data.UpdatedAt,
	}
}

func (t Tag) Data() TagData {
	return TagData{
		ID: t.id, Name: t.name, Slug: t.slug,
		CreatedAt: t.createdAt, UpdatedAt: t.updatedAt,
	}
}

func NewTag(name, slug string) (Tag, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return Tag{}, ErrTagNameRequired
	}
	if slug == "" {
		return Tag{}, ErrTagSlugRequired
	}
	return Tag{name: name, slug: slug}, nil
}

// Rename trả về thẻ mới với tên và slug mới, giữ nguyên ID và thời gian tạo.
func (t Tag) Rename(name, slug string) (Tag, error) {
	renamed, err := NewTag(name, slug)
	if err != nil {
		return Tag{}, err
	}
	renamed.id = t.id
	renamed.createdAt = t.createdAt
	renamed.updatedAt = t.updatedAt
	return renamed, nil
}

package domain

import "time"

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

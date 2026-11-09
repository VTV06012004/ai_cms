package domain

import (
	"context"
	"time"
)

type PostListCriteria struct {
	AuthorID         uint
	IncludePublished bool
	PublishedOnly    bool
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	GetAll(ctx context.Context, page PageRequest) ([]User, int64, error)
	CountByRole(ctx context.Context, role UserRole) (int64, error)
	UpdateRole(ctx context.Context, userID uint, newRole string) error
}

type PostRepository interface {
	GetAll(ctx context.Context, criteria PostListCriteria, page PageRequest) ([]Post, int64, error)
	GetByID(ctx context.Context, id uint) (*Post, error)
	Create(ctx context.Context, post *Post, tagIDs []uint) (*Post, error)
	Update(ctx context.Context, post *Post, tagIDs []uint) (*Post, error)
	Delete(ctx context.Context, id uint) error
	ListDue(ctx context.Context, now time.Time, limit int) ([]Post, error)
	PublishIfDue(ctx context.Context, id uint, now time.Time) (bool, error)
}

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id uint) (*Category, error)
	Create(ctx context.Context, category *Category) (*Category, error)
	Update(ctx context.Context, category *Category) (*Category, error)
	Delete(ctx context.Context, id uint) error
	CountPosts(ctx context.Context, id uint) (int64, error)
	CountChildren(ctx context.Context, id uint) (int64, error)
}

type TagRepository interface {
	GetAll(ctx context.Context) ([]Tag, error)
	GetByID(ctx context.Context, id uint) (*Tag, error)
	Create(ctx context.Context, tag Tag) (Tag, error)
	Update(ctx context.Context, tag Tag) (Tag, error)
	Delete(ctx context.Context, id uint) error
}

type MenuRepository interface {
	GetAll(ctx context.Context) ([]Menu, error)
	GetByID(ctx context.Context, id uint) (*Menu, error)
	Create(ctx context.Context, menu *Menu) (*Menu, error)
	Update(ctx context.Context, menu *Menu) (*Menu, error)
	Delete(ctx context.Context, id uint) error
}

type RevisionRepository interface {
	Create(ctx context.Context, revision *Revision) error
	GetByPostID(ctx context.Context, postID uint) ([]Revision, error)
	GetByID(ctx context.Context, id uint) (*Revision, error)
}

type AuditRepository interface {
	RecordLog(ctx context.Context, userID uint, action, target, details, ipAddress string) error
	GetAll(ctx context.Context, page PageRequest) ([]AuditLog, int64, error)
}

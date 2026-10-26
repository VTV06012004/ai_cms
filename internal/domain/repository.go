package domain

type PostListCriteria struct {
	AuthorID         uint
	IncludePublished bool
}

type UserRepository interface {
	Create(user *User) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByID(id uint) (*User, error)
	GetAll() ([]User, error)
	UpdateRole(userID uint, newRole string) error
}

type PostRepository interface {
	GetAll(criteria PostListCriteria) ([]Post, error)
	GetByID(id uint) (*Post, error)
	Create(post *Post, tagIDs []uint) (*Post, error)
	Update(post *Post, tagIDs []uint) (*Post, error)
	Delete(id uint) error
}

type CategoryRepository interface {
	GetAll() ([]Category, error)
	GetByID(id uint) (*Category, error)
	Create(category *Category) (*Category, error)
}

type TagRepository interface {
	GetAll() ([]Tag, error)
}

type MenuRepository interface {
	GetAll() ([]Menu, error)
	GetByID(id uint) (*Menu, error)
	Create(menu *Menu) (*Menu, error)
	Delete(id uint) error
}

type RevisionRepository interface {
	Create(revision *Revision) error
	GetByPostID(postID uint) ([]Revision, error)
	GetByID(id uint) (*Revision, error)
}

type AuditRepository interface {
	RecordLog(userID uint, action, target, details, ipAddress string) error
	GetAll() ([]AuditLog, error)
}

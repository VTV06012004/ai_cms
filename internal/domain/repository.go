package domain

type PostListCriteria struct {
	AuthorID         uint
	IncludePublished bool
	PublishedOnly    bool
}

type UserRepository interface {
	Create(user *User) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByID(id uint) (*User, error)
	GetAll(page PageRequest) ([]User, int64, error)
	CountByRole(role UserRole) (int64, error)
	UpdateRole(userID uint, newRole string) error
}

type PostRepository interface {
	GetAll(criteria PostListCriteria, page PageRequest) ([]Post, int64, error)
	GetByID(id uint) (*Post, error)
	Create(post *Post, tagIDs []uint) (*Post, error)
	Update(post *Post, tagIDs []uint) (*Post, error)
	Delete(id uint) error
}

type CategoryRepository interface {
	GetAll() ([]Category, error)
	GetByID(id uint) (*Category, error)
	Create(category *Category) (*Category, error)
	Update(category *Category) (*Category, error)
	Delete(id uint) error
	CountPosts(id uint) (int64, error)
	CountChildren(id uint) (int64, error)
}

type TagRepository interface {
	GetAll() ([]Tag, error)
	GetByID(id uint) (*Tag, error)
	Create(tag Tag) (Tag, error)
	Update(tag Tag) (Tag, error)
	Delete(id uint) error
}

type MenuRepository interface {
	GetAll() ([]Menu, error)
	GetByID(id uint) (*Menu, error)
	Create(menu *Menu) (*Menu, error)
	Update(menu *Menu) (*Menu, error)
	Delete(id uint) error
}

type RevisionRepository interface {
	Create(revision *Revision) error
	GetByPostID(postID uint) ([]Revision, error)
	GetByID(id uint) (*Revision, error)
}

type AuditRepository interface {
	RecordLog(userID uint, action, target, details, ipAddress string) error
	GetAll(page PageRequest) ([]AuditLog, int64, error)
}

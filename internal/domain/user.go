package domain

import (
	"net/mail"
	"strings"
	"time"
)

type UserRole string
type Email string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleEditor UserRole = "editor"
	UserRoleAuthor UserRole = "author"
)

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(value)
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", ErrInvalidEmail
	}
	return Email(value), nil
}

func (role UserRole) CanPublishDirectly() bool {
	return role == UserRoleAdmin || role == UserRoleEditor
}

func (role UserRole) CanSubmitPostStatus(status PostStatus) bool {
	switch status {
	case PostStatusDraft, PostStatusReview:
		return role == UserRoleAdmin || role == UserRoleEditor || role == UserRoleAuthor
	case PostStatusPublished:
		return role.CanPublishDirectly()
	default:
		return false
	}
}

type User struct {
	id           uint
	email        Email
	passwordHash string
	name         string
	role         UserRole
	canApprove   bool
	rootAdmin    bool
	createdAt    time.Time
	updatedAt    time.Time
}

type UserData struct {
	ID           uint
	Email        Email
	PasswordHash string
	Name         string
	Role         UserRole
	CanApprove   bool
	RootAdmin    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(name, email, passwordHash string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrUserNameRequired
	}
	validEmail, err := NewEmail(email)
	if err != nil {
		return nil, err
	}
	if passwordHash == "" {
		return nil, ErrPasswordHashRequired
	}
	return &User{
		name:         name,
		email:        validEmail,
		passwordHash: passwordHash,
		role:         UserRoleAuthor,
	}, nil
}

func NewRootAdminUser(name, email, passwordHash string) (*User, error) {
	user, err := NewUser(name, email, passwordHash)
	if err != nil {
		return nil, err
	}
	user.role = UserRoleAdmin
	user.rootAdmin = true
	return user, nil
}

func ReconstituteUser(data UserData) *User {
	return &User{
		id: data.ID, email: data.Email, passwordHash: data.PasswordHash,
		name: data.Name, role: data.Role, canApprove: data.CanApprove, rootAdmin: data.RootAdmin,
		createdAt: data.CreatedAt, updatedAt: data.UpdatedAt,
	}
}

func (u *User) Data() UserData {
	return UserData{
		ID: u.id, Email: u.email, PasswordHash: u.passwordHash,
		Name: u.name, Role: u.role, CanApprove: u.canApprove, RootAdmin: u.rootAdmin,
		CreatedAt: u.createdAt, UpdatedAt: u.updatedAt,
	}
}

func cloneUser(user *User) *User {
	if user == nil {
		return nil
	}
	return ReconstituteUser(user.Data())
}

func (u *User) ChangeRole(role UserRole) error {
	if u.rootAdmin {
		return ErrCannotChangeRootAdmin
	}
	switch role {
	case UserRoleAdmin, UserRoleEditor, UserRoleAuthor:
		u.role = role
		return nil
	default:
		return ErrInvalidRole
	}
}

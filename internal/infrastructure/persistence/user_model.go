package persistence

import (
	"time"

	"api/internal/domain"

	"gorm.io/gorm"
)

type UserModel struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Name         string
	Role         string `gorm:"type:enum('admin','editor','author');default:'author'"`
	CanApprove   bool   `gorm:"default:false"`
	// Nullable unique key allows multiple non-root users but only one root admin.
	RootAdminKey *uint `gorm:"uniqueIndex"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string { return "users" }

func (m UserModel) ToDomain() domain.User {
	return *userModelToDomain(m)
}

func userModelToDomain(m UserModel) *domain.User {
	isRootAdmin := m.RootAdminKey != nil && *m.RootAdminKey == 1
	return domain.ReconstituteUser(domain.UserData{
		ID:           m.ID,
		Email:        domain.Email(m.Email),
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		Role:         domain.UserRole(m.Role),
		CanApprove:   m.CanApprove,
		RootAdmin:    isRootAdmin,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	})
}

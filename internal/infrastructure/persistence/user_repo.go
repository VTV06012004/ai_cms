package persistence

import (
	"api/internal/domain"
	"context"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	data := user.Data()
	var rootAdminKey *uint
	if data.RootAdmin {
		key := uint(1)
		rootAdminKey = &key
	}
	model := UserModel{
		Name:         data.Name,
		Email:        string(data.Email),
		PasswordHash: data.PasswordHash,
		Role:         string(data.Role),
		RootAdminKey: rootAdminKey,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isDuplicateKey(err, "email") {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, err
	}
	created := model.ToDomain()
	return &created, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, notFoundOr(err, domain.ErrUserNotFound)
	}
	return userModelToDomain(user), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user UserModel
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, notFoundOr(err, domain.ErrUserNotFound)
	}
	return userModelToDomain(user), nil
}

func (r *UserRepository) GetAll(ctx context.Context, page domain.PageRequest) ([]domain.User, int64, error) {
	var total int64
	db := r.db.WithContext(ctx)
	if err := db.Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []UserModel
	err := db.Select("id", "name", "email", "role", "created_at").
		Order("id ASC").
		Limit(page.PerPage).Offset(page.Offset()).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	result := make([]domain.User, 0, len(users))
	for _, user := range users {
		result = append(result, user.ToDomain())
	}
	return result, total, nil
}

func (r *UserRepository) CountByRole(ctx context.Context, role domain.UserRole) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID uint, newRole string) error {
	db := r.db.WithContext(ctx)
	result := db.Model(&UserModel{}).Where("id = ?", userID).Update("role", newRole)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var users int64
		if err := db.Model(&UserModel{}).Where("id = ?", userID).Count(&users).Error; err != nil {
			return err
		}
		if users == 0 {
			return domain.ErrUserNotFound
		}
	}
	return nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID uint, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&UserModel{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

var _ domain.UserRepository = (*UserRepository)(nil)

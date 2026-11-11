package persistence

import (
	"context"
	"errors"
	"time"

	"api/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID uint, tokenHash string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Create(&RefreshTokenModel{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}).Error
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldHash, newHash string, now, expiresAt time.Time) (uint, error) {
	var userID uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current RefreshTokenModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND expires_at > ?", oldHash, now).
			First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrInvalidRefreshToken
		}
		if err != nil {
			return err
		}
		result := tx.Delete(&current)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return domain.ErrInvalidRefreshToken
		}
		if err := tx.Create(&RefreshTokenModel{
			TokenHash: newHash,
			UserID:    current.UserID,
			ExpiresAt: expiresAt,
		}).Error; err != nil {
			return err
		}
		userID = current.UserID
		return nil
	})
	return userID, err
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).
		Delete(&RefreshTokenModel{}).Error
}

var _ domain.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

package persistence

import "time"

type RefreshTokenModel struct {
	TokenHash string `gorm:"primaryKey;size:64"`
	UserID    uint   `gorm:"not null;index"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }

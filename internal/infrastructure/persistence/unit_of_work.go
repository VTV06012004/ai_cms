package persistence

import (
	"api/internal/usecase"
	"context"

	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *GormUnitOfWork {
	return &GormUnitOfWork{db: db}
}

func (u *GormUnitOfWork) WithinTransaction(ctx context.Context, work func(usecase.TransactionRepositories) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repositories := usecase.TransactionRepositories{
			Users:      NewUserRepository(tx),
			Posts:      NewPostRepository(tx),
			Categories: NewCategoryRepository(tx),
			Tags:       NewTagRepository(tx),
			Menus:      NewMenuRepository(tx),
			Revisions:  NewRevisionRepository(tx),
			Audits:     NewAuditRepository(tx),
		}
		return work(repositories)
	})
}

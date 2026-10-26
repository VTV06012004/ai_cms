package persistence

import (
	"api/internal/usecase"

	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *GormUnitOfWork {
	return &GormUnitOfWork{db: db}
}

func (u *GormUnitOfWork) WithinTransaction(work func(usecase.TransactionRepositories) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repositories := usecase.TransactionRepositories{
			Users:      NewUserRepository(tx),
			Posts:      NewPostRepository(tx),
			Categories: NewCategoryRepository(tx),
			Menus:      NewMenuRepository(tx),
			Revisions:  NewRevisionRepository(tx),
			Audits:     NewAuditRepository(tx),
		}
		return work(repositories)
	})
}

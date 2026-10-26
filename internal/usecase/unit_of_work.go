package usecase

import "api/internal/domain"

type TransactionRepositories struct {
	Users      domain.UserRepository
	Posts      domain.PostRepository
	Categories domain.CategoryRepository
	Menus      domain.MenuRepository
	Revisions  domain.RevisionRepository
	Audits     domain.AuditRepository
}

type UnitOfWork interface {
	WithinTransaction(func(TransactionRepositories) error) error
}

package usecase

import "api/internal/domain"

type TransactionRepositories struct {
	Users      domain.UserRepository
	Posts      domain.PostRepository
	Categories domain.CategoryRepository
	Tags       domain.TagRepository
	Menus      domain.MenuRepository
	Revisions  domain.RevisionRepository
	Audits     domain.AuditRepository
}

type UnitOfWork interface {
	WithinTransaction(func(TransactionRepositories) error) error
}

package usecase

import (
	"api/internal/domain"
	"context"
)

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
	WithinTransaction(context.Context, func(TransactionRepositories) error) error
}

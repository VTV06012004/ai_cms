package usecase

import (
	"context"
	"github.com/VTV06012004/ai_cms/internal/domain"
)

type AuditUseCase struct {
	auditRepo domain.AuditRepository
}

func NewAuditUseCase(repo domain.AuditRepository) *AuditUseCase {
	return &AuditUseCase{auditRepo: repo}
}

func (uc *AuditUseCase) GetAllAuditLogs(ctx context.Context, page domain.PageRequest) ([]domain.AuditLog, domain.PageInfo, error) {
	logs, total, err := uc.auditRepo.GetAll(ctx, page)
	if err != nil {
		return nil, domain.PageInfo{}, err
	}
	return logs, page.Info(total), nil
}

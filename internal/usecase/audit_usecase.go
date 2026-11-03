package usecase

import "api/internal/domain"

type AuditUseCase struct {
	auditRepo domain.AuditRepository
}

func NewAuditUseCase(repo domain.AuditRepository) *AuditUseCase {
	return &AuditUseCase{auditRepo: repo}
}

func (uc *AuditUseCase) GetAllAuditLogs(page domain.PageRequest) ([]domain.AuditLog, domain.PageInfo, error) {
	logs, total, err := uc.auditRepo.GetAll(page)
	if err != nil {
		return nil, domain.PageInfo{}, err
	}
	return logs, page.Info(total), nil
}

package usecase

import "api/internal/domain"

type AuditUseCase struct {
	auditRepo domain.AuditRepository
}

func NewAuditUseCase(repo domain.AuditRepository) *AuditUseCase {
	return &AuditUseCase{auditRepo: repo}
}

func (uc *AuditUseCase) GetAllAuditLogs() ([]domain.AuditLog, error) {
	return uc.auditRepo.GetAll()
}

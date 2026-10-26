package persistence

import (
	"api/internal/domain"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) RecordLog(userID uint, action, target, details, ipAddress string) error {
	data := domain.NewAuditLog(userID, action, target, details, ipAddress).Data()
	log := AuditLogModel{
		UserID:    data.UserID,
		Action:    data.Action,
		Target:    data.Target,
		Details:   data.Details,
		IPAddress: data.IPAddress,
	}
	return r.db.Create(&log).Error
}

func (r *AuditRepository) GetAll() ([]domain.AuditLog, error) {
	var logs []AuditLogModel
	if err := r.db.Preload("User").Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	var _ domain.AuditRepository = (*AuditRepository)(nil)
	result := make([]domain.AuditLog, 0, len(logs))
	for _, log := range logs {
		result = append(result, log.ToDomain())
	}
	return result, nil
}

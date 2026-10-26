package persistence

import (
	"api/internal/domain"
	"time"
)

type AuditLogModel struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	User      UserModel `gorm:"foreignKey:UserID"`
	Action    string    `gorm:"not null"`
	Target    string
	Details   string `gorm:"type:text"`
	IPAddress string
	CreatedAt time.Time
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

func (m AuditLogModel) ToDomain() domain.AuditLog {
	data := domain.AuditLogData{
		ID:        m.ID,
		UserID:    m.UserID,
		Action:    m.Action,
		Target:    m.Target,
		Details:   m.Details,
		IPAddress: m.IPAddress,
		CreatedAt: m.CreatedAt,
	}
	if m.User.ID != 0 {
		data.User = userModelToDomain(m.User)
	}
	return *domain.ReconstituteAuditLog(data)
}

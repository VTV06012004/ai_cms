package domain

import "time"

type AuditLog struct {
	id        uint
	userID    uint
	user      *User
	action    string
	target    string
	details   string
	ipAddress string
	createdAt time.Time
}

type AuditLogData struct {
	ID        uint
	UserID    uint
	User      *User
	Action    string
	Target    string
	Details   string
	IPAddress string
	CreatedAt time.Time
}

func NewAuditLog(userID uint, action, target, details, ipAddress string) *AuditLog {
	return ReconstituteAuditLog(AuditLogData{
		UserID: userID, Action: action, Target: target,
		Details: details, IPAddress: ipAddress,
	})
}

func ReconstituteAuditLog(data AuditLogData) *AuditLog {
	return &AuditLog{
		id: data.ID, userID: data.UserID, user: cloneUser(data.User),
		action: data.Action, target: data.Target, details: data.Details,
		ipAddress: data.IPAddress, createdAt: data.CreatedAt,
	}
}

func (l *AuditLog) Data() AuditLogData {
	return AuditLogData{
		ID: l.id, UserID: l.userID, User: cloneUser(l.user),
		Action: l.action, Target: l.target, Details: l.details,
		IPAddress: l.ipAddress, CreatedAt: l.createdAt,
	}
}

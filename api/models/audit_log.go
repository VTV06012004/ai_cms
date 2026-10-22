package models

import (
	"time"
)

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"` // Ai thực hiện hành động
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action    string    `gorm:"not null" json:"action"`   // Hành động (VD: LOGIN, CREATE_POST, DELETE_POST, RESTORE_REVISION)
	Target    string    `json:"target"`                   // Đối tượng tác động (VD: Post #5, User #2)
	Details   string    `gorm:"type:text" json:"details"` // Mô tả chi tiết thêm
	IPAddress string    `json:"ip_address"`               // Địa chỉ IP của người dùng
	CreatedAt time.Time `json:"created_at"`               // Thời gian diễn ra
}
